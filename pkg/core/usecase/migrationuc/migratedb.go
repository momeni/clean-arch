// Copyright (c) 2024 Behnam Momeni
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package migrationuc

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/momeni/clean-arch/pkg/core/cerr"
	"github.com/momeni/clean-arch/pkg/core/model"
	"github.com/momeni/clean-arch/pkg/core/repo"
	"gopkg.in/yaml.v3"
)

// MigrateDBUseCase represents the configuration and database schema
// migration use cases. The source and destination versions are
// determined by Settings instances which also provide the database
// connection information for the source and destination databases.
// For details, see the NewMigrateDB function.
type MigrateDBUseCase struct {
	migrator      repo.Migrator[Settings] // source configs migrator
	dstSettings   Settings                // dst def. settings & db info
	targetCfgPath string                  // target config file path
	schemaRepo    repo.Schema             // schema management repo

	// srcSettings contains the source settings which are loaded by
	// the migrator and then migrated upwards or downwards in order to
	// match with the major version of the dstSettings field.
	srcSettings Settings

	// targetSettings is a clone of srcSettings where some of its
	// entires are overridden by the dstSettings and should be written
	// into targetCfgPath file (and parts of it may be required to be
	// saved in the destination database itself).
	targetSettings Settings

	// loader is a function which can be used for reading a config file
	// and loading its contents as a Settings interface. The config file
	// internal format may belong to any supported version.
	// loader is used for reading an old version of the targetSettings
	// in order to detect an incomplete migration attempt which could
	// reach to the commitment phase (and so has filled the destination
	// database), so it can be completed by dropping extra schema and
	// moving the target configuration file.
	loader ConfigFileLoader
}

// ConfigFileLoader is a version-independent function type which accepts
// a configuration file path, reads it entirely, parses it based on its
// detected version, and finally adapts it to the Settings interface.
type ConfigFileLoader func(
	ctx context.Context, path string,
) (Settings, error)

// NewMigrateDB creates a new MigrateDBUseCase instance which uses
// the `mig` migrator in order to learn about the source configuration
// file settings. Using `mig`, source settings can be migrated upwards
// or downwards until they match with the format version which is used
// by the `dstSettings` settings. Thereafter, those items which are
// missing due to the version changes (such as fields which are recently
// introduced and had no older counterpart) and those items which must
// change unconditionally (such as the database connection information
// which must be taken from the destination configuration settings) will
// be overwritten from `dstSettings` into the settings instance which
// was obtained from the `mig` settings migrator.
// Similarly, the source and destination database schema version and
// connection information are taken from `mig` and `dstSettings` in
// order to build relevant schema migrator and migrate the source schema
// upwards or downwards until it matches with the expected destination
// schema version (only if the source and destination databases are
// different and the destination database is empty).
// The `targetCfgPath` indicates the final path which should be
// overwritten atomically when the target settings are prepared.
//
// The `repo.Schema` repository is taken from the `dstSettings` in order
// to create roles and schema and drop temporary/intermediate schema.
// See repo.Schema for more details.
//
// Target settings are first written into `targetCfgPath + ".migrated"`
// file, then destination database contents are committed, and finally
// extra schema are dropped and that migrated settings file is moved
// in order to atomically write the `targetCfgPath` file.
// The loader argument is used during an incomplete migration operation
// resumption, as it allows the target migrated configuration file to
// be read and compared with the `dstSettings`. If the destination
// database is not empty, but there is a `.migrated` file which its
// contents match with the expected destination database settings, we
// can conclude that database is filled by a previous migration attempt
// and we can resume it from the commitment step.
//
// NewMigrateDB only creates the migrator and performs no actual
// operation, hence, it may not return an error.
func NewMigrateDB(
	mig repo.Migrator[Settings],
	dstSettings Settings,
	targetCfgPath string,
	loader ConfigFileLoader,
) *MigrateDBUseCase {
	return &MigrateDBUseCase{
		migrator:      mig,
		dstSettings:   dstSettings,
		targetCfgPath: targetCfgPath,
		schemaRepo:    dstSettings.NewSchemaRepo(),
		loader:        loader,
	}
}

// Migrate runs the configuration settings and database schema migration
// determining the migration direction by the settings versions and
// schema versions as recorded in the source and destination configs.
func (mduc *MigrateDBUseCase) Migrate(ctx context.Context) error {
	resumed, err := mduc.migrateSettings(ctx)
	if err != nil {
		return fmt.Errorf("migrating settings: %w", err)
	}
	if !resumed {
		if err := mduc.migrateDBAndSaveConfigFile(ctx); err != nil {
			return fmt.Errorf("migrating db & saving settings: %w", err)
		}
	}
	if err := mduc.commitTargetSettings(); err != nil {
		return fmt.Errorf("committing target settings: %w", err)
	}
	return nil
}

// migrateSettings uses `mduc.migrator` source settings migrator for
// migrating it upwards/downwards until it matches with the version of
// the `mduc.dstSettings`. However, it does not merge it with the
// destination settings and stores it in the `mduc.srcSettings` field.
// That is, if this function returns nil error, the `mduc.srcSettings`
// will be initialized and its settings format version will match with
// the `mduc.dstSettings` and it is enough to merge the destination
// settings into the `mduc.srcSettings` fields in order to obtain the
// target settings. Therefore, the source and destination database
// schema versions can be taken from the `mduc.srcSettings` and
// `mduc.dstSettings` instances respectively.
// That remaining settings merging operation is also performed on a
// cloned version of `mduc.srcSettings` and stored in `targetSettings`
// field of `mduc`. The `targetSettings` should be written into a
// file before committing the destination database changes (and if it
// could reach to its final state with no other error).
//
// The resumed boolean return value indicates that a previous migration
// attempt had failed, while the src and dst databases were the same (so
// only the configuration files should be migrated) and the settings
// migration had proceeded enough to write a `.migrated` file and commit
// the target mutable settings in the database too. That is, resumed
// boolean informs caller that only the configuration file commitment
// step is remaining for completion of the migration.
func (mduc *MigrateDBUseCase) migrateSettings(
	ctx context.Context,
) (resumed bool, err error) {
	mig := mduc.migrator
	srcMajorVer := mig.MajorVersion()
	dstCfgVer := mduc.dstSettings.Version()
	dstMajorVer := dstCfgVer[0]
	ss, err := obtainSettler(ctx, mig, srcMajorVer, dstMajorVer)
	if err != nil {
		var msve *cerr.MismatchingSemVerError
		if ss == nil || !errors.As(err, &msve) {
			return false, fmt.Errorf(
				"migrating from %d to %v major version: %w",
				srcMajorVer, dstMajorVer, err,
			)
		}
		// src config could be loaded, but could not be overridden by DB
		err = mduc.resumeUniDBMigration(ctx, ss, msve, dstCfgVer)
		if err != nil {
			return false, fmt.Errorf(
				"examining possibility of resumption: %w", err,
			)
		}
		return true, nil
	}
	mduc.srcSettings = ss
	ts := ss.Clone()
	if err := ts.MergeSettings(mduc.dstSettings); err != nil {
		return false, fmt.Errorf("merging src/dst settings: %w", err)
	}
	mduc.targetSettings = ts
	return false, nil
}

// resumeUniDBMigration checks for following conditions:
//  1. The src and dst databases (described by srcSettings and
//     mduc.dstSettings) belong to the same database (considering their
//     connection information) and the src database version is backward
//     compatible with the dst database (so it is possible to use the src
//     database instead of the dst database in a uni-database migration
//     and only update the configuration settings format),
//  2. A .migrated file was created by an old migration attempt,
//  3. The .migrated file describes the same dst database and is
//     compatible with it (so it can be accepted as a migrated version
//     of the src settings),
//  4. The configuration settings version of the .migrated file is
//     also backward compatible with the settings version which was
//     asked by the dstCfgVer argument,
//  5. The unexpected version of the settings which were stored in the
//     src/dst database, as reported by the srcSettingsOverridingErr
//     argument, is exactly equal with the .migrated settings format
//     version (and it was committed in the src database because the
//     old migration had succeeded to proceed enough to commit its
//     transaction and now, only the last atomic move of the target
//     configuration file is remaining).
//
// If aforementioned conditions are met, a nil error will be returned.
func (mduc *MigrateDBUseCase) resumeUniDBMigration(
	ctx context.Context,
	srcSettings Settings,
	srcSettingsOverridingErr *cerr.MismatchingSemVerError,
	dstCfgVer model.SemVer,
) error {
	switch sameDBs, err := HasTheSameConnectionInfo(
		srcSettings, mduc.dstSettings,
	); {
	case err != nil:
		return fmt.Errorf("uni-db migration schema versions: %w", err)
	case !sameDBs:
		return fmt.Errorf(
			"multi-db migration: overriding src settings: %w",
			srcSettingsOverridingErr,
		)
	}
	ms, err := mduc.loadTargetSettings(ctx)
	if err != nil {
		return fmt.Errorf("loading .migrated file: %w", err)
	}
	switch sameDBs, err := HasTheSameConnectionInfo(
		ms, mduc.dstSettings,
	); {
	case err != nil:
		return fmt.Errorf(
			"comparing dst schema with .migrated file version: %w", err,
		)
	case !sameDBs:
		return errors.New("irrelevant .migrated file")
	}
	targetCfgVer := ms.Version()
	if !AreVersionsCompatible(targetCfgVer, dstCfgVer) {
		return fmt.Errorf(
			"dst config v%s may not be replaced by .migrated v%s",
			dstCfgVer, targetCfgVer,
		)
	}
	dstDBCfgVer := (*srcSettingsOverridingErr)[1]
	if targetCfgVer != dstDBCfgVer {
		return fmt.Errorf(
			"dst DB settings v%s does not match with .migrated v%s",
			dstDBCfgVer, targetCfgVer,
		)
	}
	return nil
}

// obtainSettler of S uses `mig` which should be a migrator of database
// schema or configuration settings in order to (1) load the source
// metadata, (2) obtain an upwards or downwards migrator object and
// migrate upwards/downwards until it reaches to dstMajorVer major
// version assuming that it had started from the srcMajorVer major
// version, and (3) then fetches the resulting migration settler object
// which will have the S generic type. The settlement itself is not
// performed. For more details about the migration phases, see the
// repo.Migrator generic interface documentation.
//
// The configuration settings are migrated by `migrateSettings` method
// which performs its settlement operation by merging the destination
// settings into the obtained target settings.
// The database schema settings are migrated by the `fillMigPathSchema`
// method which performs its settlement operation by creating the
// relevant tables and filling them using the views of other schema
// versions.
func obtainSettler[S any](
	ctx context.Context,
	mig repo.Migrator[S],
	srcMajorVer, dstMajorVer uint,
) (S, error) {
	var snil S
	if err := mig.Load(ctx); err != nil {
		// If the Load method could succeed partially, the Settler
		// method may return a settler object (and a nil error), so
		// it can be returned alongside the wrapped err.
		// If the Settler method returned a non-nil error, snil will
		// remain uninitialized.
		// Therefore, it is not required to check Settler error.
		snil, _ := mig.Settler(ctx)
		return snil, fmt.Errorf("Load(): %w", err)
	}
	if srcMajorVer > dstMajorVer {
		dm, err := mig.DownMigrator(ctx)
		if err != nil {
			return snil, fmt.Errorf("DownMigrator(): %w", err)
		}
		for srcMajorVer != dstMajorVer {
			dm, err = dm.MigrateDown(ctx)
			if err != nil {
				return snil, fmt.Errorf(
					"migrating downwards from major version %d: %w",
					srcMajorVer, err,
				)
			}
			srcMajorVer--
		}
		return dm.Settler(), nil
	}
	um, err := mig.UpMigrator(ctx)
	if err != nil {
		return snil, fmt.Errorf("UpMigrator(): %w", err)
	}
	for srcMajorVer != dstMajorVer {
		um, err = um.MigrateUp(ctx)
		if err != nil {
			return snil, fmt.Errorf(
				"migrating upwards from major version %d: %w",
				srcMajorVer, err,
			)
		}
		srcMajorVer++
	}
	return um.Settler(), nil
}

// migrateDBAndSaveConfigFile checks the source and destination database
// connection information. If they refer to the same database, it is not
// possible to migrate database contents. In this case, we expect their
// versions to be compatible too (their major versions must match and
// the minor version of source should be equal or greater than the
// destination database), so it can assumed that migration was
// successful without touching the source database.
// If they were two distinct databases, migrateDB method will be used
// for performing the database schema migration (or resuming an old
// incomplete migration attempt), writing the targetSettings into an
// intermediate file named after targetCfgPath with ".migrated" extra
// extension, and committing the destination database changes.
// Ultimately, migrateDBAndSaveConfigFile uses dropIntermediateSchema
// in order to drop unused schema which were created to contain the
// intermediate versions schema (containing relevant views).
// Therefore, if no error is returned, caller can safely proceed by
// moving the .migrated file and writing the targetCfgPath file.
func (mduc *MigrateDBUseCase) migrateDBAndSaveConfigFile(
	ctx context.Context,
) error {
	sameDBs, err := HasTheSameConnectionInfo(
		mduc.srcSettings, mduc.dstSettings,
	)
	if err != nil {
		return fmt.Errorf("comparing src/dst connection info: %w", err)
	}
	srcVer := mduc.srcSettings.SchemaVersion()
	dstVer := mduc.dstSettings.SchemaVersion()
	if sameDBs {
		mduc.targetSettings.SetSchemaVersion(dstVer)
		err = mduc.persistSettingsInDBAndFile(ctx, nil)
		if err != nil {
			return fmt.Errorf("persisting target settings: %w", err)
		}
		return nil
	}
	serverName := ForeignServerName(srcVer[0], srcVer[1])
	schemaNames := listSchemaNames(srcVer, dstVer)
	if err := mduc.migrateDB(ctx, serverName, schemaNames); err != nil {
		return fmt.Errorf("migrating database schema: %w", err)
	}
	// excluding the last item which must be kept
	schemaNames = schemaNames[:len(schemaNames)-1]
	if err := mduc.dropIntermediateSchema(
		ctx, serverName, schemaNames,
	); err != nil {
		return fmt.Errorf("dropping intermediate schema: %w", err)
	}
	return nil
}

// persistSettingsInDBAndFile examines the mduc.targetSettings instance,
// serializes its mutable settings, and uses the persister argument for
// persisting it in the database before trying to save the complete
// mduc.targetSettings in a configuration file.
// If the persister argument is nil, a new connection will be
// established (using the repo.NormalRole) and a new persister object
// will be created using the mduc.targetSettings.SettingsPersister
// in order to try the in-database (and then in-file) persistence again.
func (mduc *MigrateDBUseCase) persistSettingsInDBAndFile(
	ctx context.Context, persister repo.SettingsPersister,
) error {
	if persister != nil {
		ms, minb, maxb, err := mduc.targetSettings.Serialize()
		if err != nil {
			return fmt.Errorf("serializing target settings: %w", err)
		}
		err = persister.PersistSettings(ctx, ms, minb, maxb)
		if err != nil {
			return fmt.Errorf("saving mutable settings in DB: %w", err)
		}
		err = mduc.saveTargetSettings()
		if err != nil {
			return fmt.Errorf("saving .migrated config file: %w", err)
		}
		return nil
	}
	p, err := mduc.targetSettings.ConnectionPool(ctx, repo.NormalRole)
	if err != nil {
		return fmt.Errorf("creating DB pool for normal role: %w", err)
	}
	defer p.Close()
	err = p.Conn(ctx, func(ctx context.Context, c repo.Conn) error {
		return c.Tx(ctx, func(ctx context.Context, tx repo.Tx) error {
			sp, err := mduc.targetSettings.SettingsPersister(tx)
			if err != nil {
				return fmt.Errorf("creating SettingsPersister: %w", err)
			}
			return mduc.persistSettingsInDBAndFile(ctx, sp)
		})
	})
	if err != nil {
		return fmt.Errorf("transaction failed: %w", err)
	}
	return nil
}

// migrateDB uses dropAndCreateAgain method in order to drop schema
// starting from the target destination schema back to intermediate
// schema until the initial foreign schema (without cascading) ensuring
// that they either did not exist or were empty and so we are not going
// to overwrite some non-empty database contents mistakenly, and
// creates those schema in their usage order. If the database schema
// could be renewed successfully, fillMigPathSchema method will be used
// for filling them and performing the actual database migration. If
// renewal failed, we may have a committed destination database from an
// old migration attempt. This scenario is verified by calling the
// loadTargetSettings and comparing the .migrated file contents with
// the dstSettings. Anyways, in absence of errors, this function ensures
// that destination database contents are committed and .migrated file
// is also written out (before committing the database contents). So,
// caller is free to dropIntermediateSchema and continue by moving the
// target .migrated file into its ultimate path.
func (mduc *MigrateDBUseCase) migrateDB(
	ctx context.Context, serverName string, schemaNames []string,
) error {
	renewed, err := mduc.dropAndCreateAgain(
		ctx, serverName, schemaNames,
	)
	if err != nil {
		return fmt.Errorf(
			"MigrateDBUseCase.dropAndCreateAgain(%q, %v): %w",
			serverName, schemaNames, err,
		)
	}
	if !renewed {
		ms, err := mduc.loadTargetSettings(ctx)
		if err != nil {
			return fmt.Errorf("loading old .migrated config: %w", err)
		}
		sameDBs, err := HasTheSameConnectionInfo(ms, mduc.dstSettings)
		if err != nil {
			return fmt.Errorf("checking old .migrated config: %w", err)
		}
		if !sameDBs {
			return errors.New("dst database is not empty")
		}
		return nil // resuming an old incomplete migration
	}
	err = mduc.fillMigPathSchema(ctx)
	if err != nil {
		return fmt.Errorf("filling migration-path schema: %w", err)
	}
	return nil
}

// dropAndCreateAgain creates a connection to the destination database
// using the repo.AdminRole which must exist already, installs the
// postgres_fdw extension if it is not already installed (relevant .so
// file must be installed beforehand and this installation is only
// about the internal database objects creation), and drops the
// schemaNames in the reversed order, stopping at the first error.
// It drops the foreign server (if exists) at the end.
// Thereafter, it creates the repo.NormalRole if it is missing and
// renews passwords of both of the admin and normal roles in a way that
// abrupt termination does not harm database connectivity during the
// next call of ConnectionPool method. For details of this process, read
// the SchemaSettings.ConnectionPool and SchemaSettings.RenewPasswords
// methods.
//
// All schemaNames are created in their original order and their
// privileges are granted to the normal user in addition to the USAGE
// privilege on the postgres_fdw extension, so the normal user can
// create a foreign server and fill those schemaNames schema later.
// This strategy minimizes queries which must be executed by the admin
// role in the destination database. The foreign server and its user
// mapping may be dropped by this method, but are not created here.
func (mduc *MigrateDBUseCase) dropAndCreateAgain(
	ctx context.Context, serverName string, schemaNames []string,
) (renewed bool, err error) {
	p, err := mduc.dstSettings.ConnectionPool(ctx, repo.AdminRole)
	if err != nil {
		return false, fmt.Errorf("creating DB pool for admin: %w", err)
	}
	defer p.Close()
	var finalizer func() error
	renewed = true
	err = p.Conn(ctx, func(ctx context.Context, c repo.Conn) error {
		q := mduc.schemaRepo.Conn(c)
		if err := q.InstallFDWExtensionIfMissing(ctx); err != nil {
			return fmt.Errorf("installing FDW extension: %w", err)
		}
		for i := len(schemaNames) - 1; i >= 0; i-- {
			sn := schemaNames[i]
			if err := q.DropIfExists(ctx, sn); err != nil {
				if i == len(schemaNames)-1 {
					renewed = false
					return nil
				}
				return fmt.Errorf("dropping %q schema: %w", sn, err)
			}
		}
		if err := q.DropServerIfExists(ctx, serverName); err != nil {
			return fmt.Errorf(
				"dropping %q foreign server (with cascade): %w",
				serverName, err,
			)
		}
		if err := q.CreateRoleIfNotExists(
			ctx, repo.NormalRole,
		); err != nil {
			return fmt.Errorf("creating normal role: %w", err)
		}
		for _, sn := range schemaNames {
			if err := q.CreateSchema(ctx, sn); err != nil {
				return fmt.Errorf("creating %q schema: %w", sn, err)
			}
			if err := q.GrantPrivileges(
				ctx, sn, repo.NormalRole,
			); err != nil {
				return fmt.Errorf(
					"granting %q schema privs to normal role: %w",
					sn, err,
				)
			}
		}
		if err := q.SetSearchPath(
			ctx, schemaNames[len(schemaNames)-1], repo.NormalRole,
		); err != nil {
			return fmt.Errorf(
				"setting search_path of normal role to %q: %w",
				schemaNames[len(schemaNames)-1], err,
			)
		}
		if err := q.GrantFDWUsage(ctx, repo.NormalRole); err != nil {
			return fmt.Errorf(
				"granting FDW USAGE priv to normal role: %w", err,
			)
		}
		return c.Tx(ctx, func(ctx context.Context, tx repo.Tx) error {
			q := mduc.schemaRepo.Tx(tx)
			finalizer, err = mduc.dstSettings.RenewPasswords(
				ctx, q.ChangePasswords, repo.AdminRole, repo.NormalRole,
			)
			if err != nil {
				return fmt.Errorf("RenewPasswords: %w", err)
			}
			return nil
		})
	})
	if err != nil {
		return false, fmt.Errorf("admin connection: %w", err)
	}
	if !renewed {
		return false, nil
	}
	if err := finalizer(); err != nil {
		return false, fmt.Errorf("finalizing passwords renewal: %w", err)
	}
	return true, nil
}

// fillMigPathSchema creates a connection to the destination database
// using the repo.NormalRole and in a transaction, creates a foreign
// server representing the source database via creation of its
// SchemaMigrator object and calling its Load method. It continues
// the upwards/downwards migration using the obtainSettler method.
// Finally, it settles the migration using the SettleSchema method of
// the obtained SchemaSettler instance. In absence of errors and right
// before committing the transaction, saveTargetSettings method is used
// for writing the .migrated file (used for possible resumption).
func (mduc *MigrateDBUseCase) fillMigPathSchema(
	ctx context.Context,
) error {
	p, err := mduc.dstSettings.ConnectionPool(ctx, repo.NormalRole)
	if err != nil {
		return fmt.Errorf("creating DB pool for normal role: %w", err)
	}
	defer p.Close()
	err = p.Conn(ctx, func(ctx context.Context, c repo.Conn) error {
		return c.Tx(ctx, func(ctx context.Context, tx repo.Tx) error {
			smig, err := mduc.srcSettings.SchemaMigrator(tx)
			if err != nil {
				return fmt.Errorf("creating schema migrator: %w", err)
			}
			srcMajorVer := smig.MajorVersion()
			dstSchemaVer := mduc.dstSettings.SchemaVersion()
			dstMajorVer := dstSchemaVer[0]
			ss, err := obtainSettler(
				ctx, smig, srcMajorVer, dstMajorVer,
			)
			if err != nil {
				return fmt.Errorf(
					"migrating from %d to %v major version: %w",
					srcMajorVer, dstMajorVer, err,
				)
			}
			err = ss.SettleSchema(ctx)
			if err != nil {
				return fmt.Errorf("SettleSchema(): %w", err)
			}
			err = mduc.persistSettingsInDBAndFile(ctx, ss)
			if err != nil {
				return fmt.Errorf("persisting target settings: %w", err)
			}
			return nil
		})
	})
	if err != nil {
		return fmt.Errorf("normal connection: %w", err)
	}
	return nil
}

// dropIntermediateSchema drops schemaNames in the reversed order
// using the repo.AdminRole from the destination database before
// dropping the foreign server (if they exist).
//
// Note that the target destination schema which its name was normally
// the last item of schemaNames slice must not be included in the
// sub-slice which is passed to dropIntermediateSchema because it must
// be kept in the destination database and only its predecessors which
// were intermediary should be dropped.
func (mduc *MigrateDBUseCase) dropIntermediateSchema(
	ctx context.Context, serverName string, schemaNames []string,
) error {
	p, err := mduc.dstSettings.ConnectionPool(ctx, repo.AdminRole)
	if err != nil {
		return fmt.Errorf("creating DB pool for admin: %w", err)
	}
	defer p.Close()
	err = p.Conn(ctx, func(ctx context.Context, c repo.Conn) error {
		q := mduc.schemaRepo.Conn(c)
		for i := len(schemaNames) - 1; i >= 0; i-- {
			sn := schemaNames[i]
			if err := q.DropCascade(ctx, sn); err != nil {
				return fmt.Errorf("dropping %q schema: %w", sn, err)
			}
		}
		if err := q.DropServerIfExists(ctx, serverName); err != nil {
			return fmt.Errorf(
				"dropping %q foreign server (with cascade): %w",
				serverName, err,
			)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("admin connection: %w", err)
	}
	return nil
}

// saveTargetSettings serializes the targetSettings as a YAML file
// contents and saves it into the targetCfgPath file with an extra
// ".migrated" extension, so it is not an issue that file may be written
// out chunk by chunk as its contents are not used until they are
// completely written out, followed by a successful database commitment
// which is itself followed by an atomic file system level move
// operation, transferring this .migrated file over the targetCfgPath
// file ultimately (see the commitTargetSettings for this movement).
func (mduc *MigrateDBUseCase) saveTargetSettings() error {
	b, err := yaml.Marshal(mduc.targetSettings)
	if err != nil {
		return fmt.Errorf("marshaling to YAML: %w", err)
	}
	p := mduc.targetCfgPath + ".migrated"
	err = os.WriteFile(p, b, 0o644)
	if err != nil {
		return fmt.Errorf("writing to %q file: %w", p, err)
	}
	return nil
}

// loadTargetSettings reads the `targetCfgPath + ".migrated"` file,
// desrializes its contents based on its detected format, and converts
// them to the version-independent Settings interface. The actual
// loading is performed by a function which is implemented in the
// adapter layer, while it is taken as a version and format independent
// loader function type in this (use cases) layer.
func (mduc *MigrateDBUseCase) loadTargetSettings(
	ctx context.Context,
) (Settings, error) {
	return mduc.loader(ctx, mduc.targetCfgPath+".migrated")
}

// commitTargetSettings moves the .migrated target file into the
// expected targetCfgPath file in order to atomically mark the end of
// this migration use case.
// It is moved to a separate method in order to have its implementation
// near the saveTargetSettings and loadTargetSettings methods, so the
// correctness of target files naming and contents handling can be
// reviewed easier (increasing their maintainability).
func (mduc *MigrateDBUseCase) commitTargetSettings() error {
	return os.Rename(
		mduc.targetCfgPath+".migrated",
		mduc.targetCfgPath,
	)
}

// listSchemaNames returns the list of schema names which should be
// created in turn for importing source database schema, see the
// ForeignSchemaName function, followed by intermediate schema for
// converting the schema format one major version at a time, see
// the MigrationSchemaName function, terminated by the destination
// schema name itself, see SchemaName function.
//
// Migrating upwards may lead to creation of tables and columns which
// had no corresponding data in previous versions, filled by default
// data or left empty until users start using those new features,
// while migrating downwards may lead to loss of information as some
// data may have no room for their storage in versions which had not
// introduced the relevant tables and columns yet.
// Therefore, although migrating upwards or downwards provides the
// most accurate information in the target version, but migrating
// upwards/downwards in a cycle is not supposed to be lossless.
func listSchemaNames(srcVer, dstVer model.SemVer) []string {
	s, d := srcVer[0], dstVer[0]
	minor := srcVer[1]
	step := uint(1)
	if s > d {
		step -= 2
	}
	names := make([]string, 0, step*(d-s)+3)
	names = append(names, ForeignSchemaName(s, minor))
	for i := s; i != d; i += step {
		names = append(names, MigrationSchemaName(i))
	}
	names = append(names, MigrationSchemaName(d))
	names = append(names, SchemaName(d))
	return names
}

// ForeignServerName returns the name of a foreign server which should
// be created in order to represent the source database with the given
// major and minor version numbers within the destination database.
func ForeignServerName(major, minor uint) string {
	return fmt.Sprintf("fps%d_%d", major, minor)
}

// ForeignSchemaName returns the name of a schema which should be
// created in the destination database in order to be filled with the
// foreign tables which are imported from a foreign server which
// represents the source database with the given major and minor
// versions. When a foreign server with the name computed by the
// ForeignServerName was created and its corresponding schema was
// imported into a local schema with the name computed by the
// ForeignSchemaName, source database contents can be read from the
// destination database and used in the migration process (without
// having to pass them through this Golang process memory).
func ForeignSchemaName(major, minor uint) string {
	return fmt.Sprintf("fdw%d_%d", major, minor)
}

// MigrationSchemaName returns the intermediate schema name which is
// used for storage of a database schema with given major version
// during a migration operation.
// After loading the source database schema in a local schema which its
// name is computed by ForeignSchemaName function, it will be migrated
// upwards to its latest supported minor version (even if it is at the
// latest minor version already), filling MigrationSchemaName schema
// for the source major version. Thereafter, its schema will be used
// in order to fill MigrationSchemaName for one major version upper or
// downer based on the migration direction until it reaches to the
// destination schema major version. Finally, the last intermediate
// migration schema will be used for creation of tables in the
// destination schema which its name is computed by SchemaName function.
// The intermediate schema will be dropped at the end.
func MigrationSchemaName(major uint) string {
	return fmt.Sprintf("mig%d", major)
}
