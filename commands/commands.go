package commands

import (
	"errors"
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-client-go/utils/log"
	"io/fs"
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

type migrationArgs struct {
	url   string
	token string
	plan  string
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
		return err
	}

	conf, err := prepareConfiguration(c, migrateToRtfs)
	if err != nil {
		return err
	}

	log.Info("Invoking command")
	cmd := exec.Command("java", "-jar", file, conf.url, conf.plan, conf.token, "false")
	combinedOutput, err := cmd.CombinedOutput()
	log.Info(string(combinedOutput))
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

	files, err := findFile(dir)

	if len(files) != 1 {
		return "", errors.New("Should've found only 1 jar files, however, found: " + strconv.Itoa(len(files)))
	}

	file := files[0]
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

func prepareConfiguration(c *components.Context, migrateToRtfs bool) (*migrationArgs, error) {
	var conf = new(migrationArgs)

	err := prepareUrl(c, conf)
	if err != nil {
		return nil, err
	}
	err = prepareToken(c, conf)
	if err != nil {
		return nil, err
	}

	preparePlan(migrateToRtfs, conf)
	log.Info("Using plan " + conf.plan)
	log.Info(conf.url + " " + conf.token)
	return conf, err
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

func findFile(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(pathFile string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			if strings.HasSuffix(d.Name(), "jar") && strings.HasPrefix(d.Name(), "on-prem") {
				files = append(files, pathFile)
			}
		}
		return nil
	})
	return files, err
}
