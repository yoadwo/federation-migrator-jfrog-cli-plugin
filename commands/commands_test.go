package commands

import (
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPrepareConfigurationRtToRTFS(t *testing.T) {
	url := "http://localhost:8081"
	token := "token"
	strings := []string{url, token}
	context := &components.Context{Arguments: strings}
	configuration, err := prepareConfiguration(context, true)
	assert.NoError(t, err)
	assert.Equal(t, configuration.url, url)
	assert.Equal(t, configuration.token, token)
	assert.Equal(t, configuration.plan, "RT_TO_RTFS")
}

func TestPrepareConfigurationRtToRTFSURLWithSlashArtifactory(t *testing.T) {
	url := "http://localhost:8081"
	token := "token"
	strings := []string{url + "/artifactory", token}
	context := &components.Context{Arguments: strings}
	configuration, err := prepareConfiguration(context, true)
	assert.NoError(t, err)
	assert.Equal(t, configuration.url, url)
	assert.Equal(t, configuration.token, token)
	assert.Equal(t, configuration.plan, "RT_TO_RTFS")
}

func TestWithEmptyToken(t *testing.T) {
	url := "http://localhost:8081"
	token := ""
	strings := []string{url + "/artifactory", token}
	context := &components.Context{Arguments: strings}
	_, err := prepareConfiguration(context, true)
	assert.EqualError(t, err, "no token provided")
}

func TestWithEmptyUrl(t *testing.T) {
	url := ""
	token := "token"
	strings := []string{url, token}
	context := &components.Context{Arguments: strings}
	_, err := prepareConfiguration(context, true)
	assert.EqualError(t, err, "need to provide url")
}

func TestPrepareConfigurationRTFSToRT(t *testing.T) {
	url := "http://localhost:8081"
	token := "token"
	strings := []string{url, token}
	context := &components.Context{Arguments: strings}
	configuration, err := prepareConfiguration(context, false)
	assert.NoError(t, err)
	assert.Equal(t, configuration.url, url)
	assert.Equal(t, configuration.token, token)
	assert.Equal(t, configuration.plan, "RTFS_TO_RT")
}

func TestMigrateWithContextNoTwoArgs(t *testing.T) {
	url := "http://localhost:8081"
	strings := []string{url}
	context := &components.Context{Arguments: strings}
	err := migrate(context, true)
	assert.EqualError(t, err, "Need to provide two arguments while provided 1")
}

func TestMigrateNoDir(t *testing.T) {
	url := "http://localhost:8081"
	token := "token"
	strings := []string{url, token}
	context := &components.Context{Arguments: strings}
	err := migrate(context, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "doesn't exists")
}

func TestPrepareReplicationToFederationConfiguration(t *testing.T) {
	url := "http://localhost:8081"
	token := "token"
	targetUrl := "http://target:8081"
	targetToken := "target-token"
	context := &components.Context{Arguments: []string{url, token, targetUrl, targetToken}}
	configuration, err := prepareReplicationToFederationConfiguration(context)
	assert.NoError(t, err)
	assert.Equal(t, url, configuration.url)
	assert.Equal(t, token, configuration.token)
	assert.Equal(t, targetUrl, configuration.targetUrl)
	assert.Equal(t, targetToken, configuration.targetToken)
	assert.Equal(t, "REPLICATION_TO_FEDERATION", configuration.plan)
}

func TestPrepareReplicationToFederationConfigurationURLWithSlashArtifactory(t *testing.T) {
	url := "http://localhost:8081"
	context := &components.Context{Arguments: []string{url + "/artifactory", "token", "http://target:8081/artifactory", "target-token"}}
	configuration, err := prepareReplicationToFederationConfiguration(context)
	assert.NoError(t, err)
	assert.Equal(t, url, configuration.url)
	assert.Equal(t, "http://target:8081", configuration.targetUrl)
}

func TestReplicationToFederationWithEmptyToken(t *testing.T) {
	context := &components.Context{Arguments: []string{"http://localhost:8081", "", "http://target:8081", "target-token"}}
	_, err := prepareReplicationToFederationConfiguration(context)
	assert.EqualError(t, err, "no token provided")
}

func TestReplicationToFederationWithEmptyUrl(t *testing.T) {
	context := &components.Context{Arguments: []string{"", "token", "http://target:8081", "target-token"}}
	_, err := prepareReplicationToFederationConfiguration(context)
	assert.EqualError(t, err, "need to provide url")
}

func TestReplicationToFederationWithEmptyTargetUrl(t *testing.T) {
	context := &components.Context{Arguments: []string{"http://localhost:8081", "token", "", "target-token"}}
	_, err := prepareReplicationToFederationConfiguration(context)
	assert.EqualError(t, err, "no target-url provided")
}

func TestReplicationToFederationWithEmptyTargetToken(t *testing.T) {
	context := &components.Context{Arguments: []string{"http://localhost:8081", "token", "http://target:8081", ""}}
	_, err := prepareReplicationToFederationConfiguration(context)
	assert.EqualError(t, err, "no target-token provided")
}

func TestMigrateReplicationToFederationWrongArgCount(t *testing.T) {
	context := &components.Context{Arguments: []string{"http://localhost:8081"}}
	err := migrateReplicationToFederation(context)
	assert.EqualError(t, err, "Need to provide four arguments while provided 1")
}

func TestMigrateReplicationToFederationNoDir(t *testing.T) {
	context := &components.Context{Arguments: []string{"http://localhost:8081", "token", "http://target:8081", "target-token"}}
	err := migrateReplicationToFederation(context)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "doesn't exists")
}

func TestBuildReplicationToFederationCommandArgs(t *testing.T) {
	conf := &migrationArgs{
		url:         "http://source:8081",
		token:       "src-token",
		plan:        "REPLICATION_TO_FEDERATION",
		targetUrl:   "http://target:8081",
		targetToken: "tgt-token",
	}
	args := buildReplicationToFederationCommandArgs("migration.jar", conf)
	assert.Contains(t, args, "-tu")
	assert.Contains(t, args, "http://target:8081")
	assert.Contains(t, args, "-tt")
	assert.Contains(t, args, "tgt-token")
	assert.Contains(t, args, "REPLICATION_TO_FEDERATION")
	assert.NotContains(t, args, "-amr")
	assert.NotContains(t, args, "-bd")
}

func TestBuildReplicationToFederationCommandArgsWithOptionalFlags(t *testing.T) {
	conf := &migrationArgs{
		url:                       "http://source:8081",
		token:                     "src-token",
		plan:                      "REPLICATION_TO_FEDERATION",
		targetUrl:                 "http://target:8081",
		targetToken:               "tgt-token",
		allowMultipleReplications: true,
		repos:                     "repo1,repo2",
		outputFile:                "/tmp/results.txt",
		conversionPollInterval:    3000,
		bidirectional:             true,
	}
	args := buildReplicationToFederationCommandArgs("migration.jar", conf)
	assert.Contains(t, args, "-amr")
	assert.Contains(t, args, "-r")
	assert.Contains(t, args, "repo1,repo2")
	assert.Contains(t, args, "-o")
	assert.Contains(t, args, "/tmp/results.txt")
	assert.Contains(t, args, "-cpi")
	assert.Contains(t, args, "3000")
	assert.Contains(t, args, "-bd")
}

func TestBuildReplicationToFederationCommandArgsWithRepoListFile(t *testing.T) {
	conf := &migrationArgs{
		url:         "http://source:8081",
		token:       "src-token",
		plan:        "REPLICATION_TO_FEDERATION",
		targetUrl:   "http://target:8081",
		targetToken: "tgt-token",
		repoListFile: "/path/to/repos.txt",
	}
	args := buildReplicationToFederationCommandArgs("migration.jar", conf)
	assert.Contains(t, args, "-rlf")
	assert.Contains(t, args, "/path/to/repos.txt")
	assert.NotContains(t, args, "-r")
}

func TestBuildJavaCommandArgs(t *testing.T) {
	conf := &migrationArgs{
		url:   "http://localhost:8081",
		token: "my-token",
		plan:  "RT_TO_RTFS",
	}
	args := buildJavaCommandArgs("migration.jar", conf)
	assert.Contains(t, args, "-jar")
	assert.Contains(t, args, "migration.jar")
	assert.Contains(t, args, "http://localhost:8081")
	assert.Contains(t, args, "RT_TO_RTFS")
	assert.Contains(t, args, "my-token")
	assert.NotContains(t, args, "-f")
	assert.NotContains(t, args, "-p")
	assert.NotContains(t, args, "-sr")
	assert.NotContains(t, args, "-rlcp")
	assert.NotContains(t, args, "-hvm")
}

func TestBuildJavaCommandArgsWithOptionalFlags(t *testing.T) {
	conf := &migrationArgs{
		url:                       "http://localhost:8081",
		token:                     "my-token",
		plan:                      "RT_TO_RTFS",
		force:                     true,
		parallelMode:              true,
		importBatchesSize:         500,
		statefulRun:               true,
		rtfsLegacyContextPathMode: true,
		verboseMode:               true,
		socketTimeoutMs:           60000,
		maxTotalConnections:       100,
		maxConnectionsPerRoute:    50,
		connectionPoolTtlSec:      30,
		retryCount:                3,
		executorTimeoutMin:        60,
		executorThreads:           10,
	}
	args := buildJavaCommandArgs("migration.jar", conf)
	assert.Contains(t, args, "-f")
	assert.Contains(t, args, "-p")
	assert.Contains(t, args, "-sr")
	assert.Contains(t, args, "-rlcp")
	assert.Contains(t, args, "-hvm")
	assert.Contains(t, args, "-bs")
	assert.Contains(t, args, "500")
	assert.Contains(t, args, "-hst")
	assert.Contains(t, args, "60000")
	assert.Contains(t, args, "-htc")
	assert.Contains(t, args, "100")
	assert.Contains(t, args, "-hcr")
	assert.Contains(t, args, "50")
	assert.Contains(t, args, "-hpt")
	assert.Contains(t, args, "30")
	assert.Contains(t, args, "-hrc")
	assert.Contains(t, args, "3")
	assert.Contains(t, args, "-etm")
	assert.Contains(t, args, "60")
	assert.Contains(t, args, "-et")
	assert.Contains(t, args, "10")
}

func TestMaskTokenShort(t *testing.T) {
	assert.Equal(t, "****", maskToken("short"))
	assert.Equal(t, "****", maskToken("12345678"))
	assert.Equal(t, "****", maskToken(""))
}

func TestMaskTokenLong(t *testing.T) {
	assert.Equal(t, "abcd...wxyz", maskToken("abcdefghijklmnopwxyz"))
}
