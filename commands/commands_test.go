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
	assert.Contains(t, err.Error(), " doesn't exists")
}
