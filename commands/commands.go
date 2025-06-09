package commands

import (
	"errors"
	"fmt"
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"migration-plugin/flags"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func GetMigrateToRtfsCommand() components.Command {
	return components.Command{
		Name:        "migrate_rtfs",
		Description: "Migrate Legacy federation to rtfs",
		Aliases:     []string{"mi_rtfs"},
		Arguments:   getDefaultMigrationArguments(),
		Flags:       getMigrationFlags(),
		Action: func(c *components.Context) error {
			return migrateToRTFS(c)
		},
	}
}

func GetMigrateToRTCommand() components.Command {
	return components.Command{
		Name:        "migrate_rt",
		Description: "Migrate rtfs to legacy federation",
		Aliases:     []string{"mi_rt"},
		Arguments:   getDefaultMigrationArguments(),
		Flags:       getMigrationFlags(),
		Action: func(c *components.Context) error {
			return migrateToRT(c)
		},
	}
}

func getDefaultMigrationArguments() []components.Argument {
	return []components.Argument{
		{
			Name:        "url",
			Description: "The base url without /artifactory.",
		}, {
			Name:        "token",
			Description: "The access token to use.",
		},
	}
}

func getMigrationFlags() []components.Flag {
	return []components.Flag{
		components.NewBoolFlag(flags.Force,
			"Force a migration without properly processing all events from queues",
			components.WithBoolDefaultValue(false)),
		components.NewBoolFlag(flags.Parallel,
			"Enable parallel mode for faster queue migration",
			components.WithBoolDefaultValue(false)),
		components.NewStringFlag(flags.BatchSize,
			"Batch size for RTFS import operations, larger sizes may cause performance issues",
			components.WithIntDefaultValue(250)),
		components.NewBoolFlag(flags.StatefulRun,
			"Enable stateful run that will migrate members that were not migrated in the previous run",
			components.WithBoolDefaultValue(false)),
		components.NewBoolFlag(flags.RtfsLegacyContextPath,
			"Use the legacy context path for RTFS, including the '/artifactory/service' prefix",
			components.WithBoolDefaultValue(false)),
		components.NewStringFlag(flags.HttpSocketTimeoutMs,
			"Socket timeout in milliseconds",
			components.WithIntDefaultValue(30*60*1000)), // 30 minutes
		components.NewStringFlag(flags.HttpMaxTotalConnections,
			"Maximum total HTTP client connections",
			components.WithIntDefaultValue(200)),
		components.NewStringFlag(flags.HttpMaxConnectionsPerRoute,
			"Maximum HTTP client connections per route",
			components.WithIntDefaultValue(200)),
		components.NewStringFlag(flags.HttpConnectionPoolTtlSec,
			"HTTP client connection pool TTL in seconds",
			components.WithIntDefaultValue(60)),
		components.NewStringFlag(flags.HttpRetryCount,
			"HTTP client retry count",
			components.WithIntDefaultValue(5)),
		components.NewBoolFlag(flags.HttpVerboseMode,
			"Enable verbose HTTP client mode",
			components.WithBoolDefaultValue(false)),
		components.NewStringFlag(flags.ExecutorTimeoutMin,
			"Executor timeout in minutes",
			components.WithIntDefaultValue(120)),
		components.NewStringFlag(flags.ExecutorThreads,
			"Number of executor threads",
			components.WithIntDefaultValue(200)),
	}
}

type migrationArgs struct {
	url                       string
	token                     string
	plan                      string
	force                     bool
	parallelMode              bool
	importBatchesSize         int
	statefulRun               bool
	rtfsLegacyContextPathMode bool
	socketTimeoutMs           int
	maxTotalConnections       int
	maxConnectionsPerRoute    int
	connectionPoolTtlSec      int
	retryCount                int
	verboseMode               bool
	executorTimeoutMin        int
	executorThreads           int
}

func migrateToRT(c *components.Context) error {
	return migrate(c, false)
}

func migrateToRTFS(c *components.Context) error {
	return migrate(c, true)
}

func migrate(c *components.Context, migrateToRtfs bool) error {
	if len(c.Arguments) != 2 {
		return errors.New("Need to provide two arguments while provided " + strconv.Itoa(len(c.Arguments)))
	}
	file, err := getMigrationJarFile()
	if err != nil {
		log.Error("Failed to get migration JAR file: " + err.Error())
		return err
	}
	log.Info("Using JAR file: " + file)

	conf, err := prepareConfiguration(c, migrateToRtfs)
	if err != nil {
		log.Error("Failed to prepare configuration: " + err.Error())
		return err
	}

	args := buildJavaCommandArgs(file, conf)

	// Log the complete command for debugging
	cmdStr := "java " + strings.Join(args, " ")
	log.Info("Executing command: " + cmdStr)

	cmd := exec.Command("java", args...)
	combinedOutput, err := cmd.CombinedOutput()
	log.Info("Command output:\n" + string(combinedOutput))

	if err != nil {
		log.Error("Command execution failed: " + err.Error())
	} else {
		log.Info("Command executed successfully")
	}

	return err
}

func getMigrationJarFile() (string, error) {
	dir, err := coreutils.GetJfrogPluginsResourcesDir("federation-migrator")
	if err != nil {
		return "", err
	}

	dirExists, err := exists(dir)

	if err != nil {
		return "", errors.New("Failed to check if dir " + dir + " exists.")
	}

	if !dirExists {
		return "", errors.New("Dir " + dir + " doesn't exists")
	}

	file, err := findFile(dir, "on-prem-2.0-jar-with-dependencies.jar")
	return file, err
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func findFile(root string, fileName string) (string, error) {
	filePath := filepath.Join(root, fileName)
	if _, err := os.Stat(filePath); err == nil {
		return filePath, nil
	} else if errors.Is(err, os.ErrNotExist) {
		return "", err
	} else {
		return "", errors.New("Error checking file existence: " + err.Error())
	}
}

func prepareConfiguration(c *components.Context, migrateToRtfs bool) (*migrationArgs, error) {
	conf := new(migrationArgs)

	if err := prepareUrl(c, conf); err != nil {
		return nil, err
	}

	if err := prepareToken(c, conf); err != nil {
		return nil, err
	}

	preparePlan(migrateToRtfs, conf)

	if err := extractFlagValues(c, conf); err != nil {
		return nil, err
	}

	// Log basic configuration
	log.Info("Using plan: " + conf.plan)
	log.Info("URL: " + conf.url)
	log.Info("Token: " + maskToken(conf.token))

	// Log all flag values
	logFlagValues(conf)
	return conf, nil
}

func maskToken(token string) string {
	if len(token) <= 8 {
		return "****"
	}
	return token[:4] + "..." + token[len(token)-4:]
}

func logFlagValues(conf *migrationArgs) {
	log.Info("=== Configuration Values ===")
	log.Info(fmt.Sprintf("force: %t", conf.force))
	log.Info(fmt.Sprintf("parallelMode: %t", conf.parallelMode))
	log.Info(fmt.Sprintf("importBatchesSize: %d", conf.importBatchesSize))
	log.Info(fmt.Sprintf("statefulRun: %t", conf.statefulRun))
	log.Info(fmt.Sprintf("rtfsLegacyContextPathMode: %t", conf.rtfsLegacyContextPathMode))
	log.Info(fmt.Sprintf("socketTimeoutMs: %d", conf.socketTimeoutMs))
	log.Info(fmt.Sprintf("maxTotalConnections: %d", conf.maxTotalConnections))
	log.Info(fmt.Sprintf("maxConnectionsPerRoute: %d", conf.maxConnectionsPerRoute))
	log.Info(fmt.Sprintf("connectionPoolTtlSec: %d", conf.connectionPoolTtlSec))
	log.Info(fmt.Sprintf("retryCount: %d", conf.retryCount))
	log.Info(fmt.Sprintf("verboseMode: %t", conf.verboseMode))
	log.Info(fmt.Sprintf("executorTimeoutMin: %d", conf.executorTimeoutMin))
	log.Info(fmt.Sprintf("executorThreads: %d", conf.executorThreads))
	log.Info("==========================")
}

func preparePlan(migrateToRtfs bool, conf *migrationArgs) {
	if migrateToRtfs {
		conf.plan = "RT_TO_RTFS"
	} else {
		conf.plan = "RTFS_TO_RT"
	}
}

func prepareToken(c *components.Context, conf *migrationArgs) error {
	conf.token = c.Arguments[1]
	if conf.token == "" {
		return errors.New("no token provided")
	}
	return nil
}

func prepareUrl(c *components.Context, conf *migrationArgs) error {
	url := c.Arguments[0]
	if url == "" {
		return errors.New("need to provide url")
	}

	if strings.HasSuffix(url, "/artifactory") {
		url = strings.TrimSuffix(url, "/artifactory")
	}

	conf.url = url
	return nil
}

func extractFlagValues(c *components.Context, conf *migrationArgs) error {
	if c.IsFlagSet(flags.Force) {
		conf.force = c.GetBoolFlagValue(flags.Force)
	}
	if c.IsFlagSet(flags.Parallel) {
		conf.parallelMode = c.GetBoolFlagValue(flags.Parallel)
	}
	if c.IsFlagSet(flags.StatefulRun) {
		conf.statefulRun = c.GetBoolFlagValue(flags.StatefulRun)
	}
	if c.IsFlagSet(flags.RtfsLegacyContextPath) {
		conf.rtfsLegacyContextPathMode = c.GetBoolFlagValue(flags.RtfsLegacyContextPath)
	}
	if c.IsFlagSet(flags.HttpVerboseMode) {
		conf.verboseMode = c.GetBoolFlagValue(flags.HttpVerboseMode)
	}

	intFlags := map[string]*int{
		flags.BatchSize:                  &conf.importBatchesSize,
		flags.HttpSocketTimeoutMs:        &conf.socketTimeoutMs,
		flags.HttpMaxTotalConnections:    &conf.maxTotalConnections,
		flags.HttpMaxConnectionsPerRoute: &conf.maxConnectionsPerRoute,
		flags.HttpConnectionPoolTtlSec:   &conf.connectionPoolTtlSec,
		flags.HttpRetryCount:             &conf.retryCount,
		flags.ExecutorTimeoutMin:         &conf.executorTimeoutMin,
		flags.ExecutorThreads:            &conf.executorThreads,
	}

	for flagName, valuePtr := range intFlags {
		if c.IsFlagSet(flagName) {
			value, err := c.GetIntFlagValue(flagName)
			if err != nil {
				return fmt.Errorf("invalid %s value: %w", flagName, err)
			}
			*valuePtr = value
		}
	}

	return nil
}

func buildJavaCommandArgs(jarFile string, conf *migrationArgs) []string {
	// Base command arguments
	args := []string{"-jar", jarFile, conf.url, conf.plan, conf.token}

	// Add all optional flags
	if conf.force {
		args = append(args, "-f")
	}
	if conf.parallelMode {
		args = append(args, "-p")
	}

	// Add numeric arguments
	args = append(args, "-bs", strconv.Itoa(conf.importBatchesSize))
	args = append(args, "-hst", strconv.Itoa(conf.socketTimeoutMs))
	args = append(args, "-htc", strconv.Itoa(conf.maxTotalConnections))
	args = append(args, "-hcr", strconv.Itoa(conf.maxConnectionsPerRoute))
	args = append(args, "-hpt", strconv.Itoa(conf.connectionPoolTtlSec))
	args = append(args, "-hrc", strconv.Itoa(conf.retryCount))
	args = append(args, "-etm", strconv.Itoa(conf.executorTimeoutMin))
	args = append(args, "-et", strconv.Itoa(conf.executorThreads))

	// Add boolean flags
	if conf.statefulRun {
		args = append(args, "-sr")
	}
	if conf.rtfsLegacyContextPathMode {
		args = append(args, "-rlcp")
	}
	if conf.verboseMode {
		args = append(args, "-hvm")
	}

	return args
}
