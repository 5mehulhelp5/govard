package remote

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"govard/internal/engine"
)

// Magento1Environment holds DB credentials extracted remotely from a Magento 1 local.xml.
type Magento1Environment struct {
	DB       MagentoDBInfo
	CryptKey string
}

// ProbeMagento1Environment SSHs to the remote environment and reads the local.xml via PHP
// to extract DB connection credentials. Returns Magento1Environment with filled DB fields.
func ProbeMagento1Environment(remoteName string, remoteCfg engine.RemoteConfig) (Magento1Environment, error) {
	remoteCommand := buildMagentoRemoteCommand(remoteCfg.Path, `php -r `+engine.ShellQuote(magento1LocalXMLProbePHP))
	encoded, err := runRemoteCapture(remoteName, remoteCfg, remoteCommand)
	if err != nil {
		return Magento1Environment{}, err
	}
	return decodeMagento1EnvironmentPayload(encoded)
}

func decodeMagento1EnvironmentPayload(encoded string) (Magento1Environment, error) {
	trimmed := strings.TrimSpace(encoded)
	if trimmed == "" {
		return Magento1Environment{}, fmt.Errorf("remote probe returned empty payload")
	}

	decoded, err := base64.StdEncoding.DecodeString(trimmed)
	if err != nil {
		return Magento1Environment{}, fmt.Errorf("decode remote probe payload: %w", err)
	}

	var payload struct {
		Host     string `json:"host"`
		Username string `json:"username"`
		Password string `json:"password"`
		DBName   string `json:"dbname"`
		CryptKey string `json:"crypt_key"`
	}
	if err := json.Unmarshal(decoded, &payload); err != nil {
		return Magento1Environment{}, fmt.Errorf("parse remote probe payload: %w", err)
	}

	host, port := ParseMagentoDBHostPort(payload.Host)
	username := strings.TrimSpace(payload.Username)
	database := strings.TrimSpace(payload.DBName)
	if username == "" || database == "" {
		return Magento1Environment{}, fmt.Errorf("remote local.xml is missing db username or dbname")
	}

	return Magento1Environment{
		DB: MagentoDBInfo{
			Host:     host,
			Port:     port,
			Username: username,
			Password: payload.Password,
			Database: database,
		},
		CryptKey: strings.TrimSpace(payload.CryptKey),
	}, nil
}

// magento1LocalXMLProbePHP is a small PHP snippet that reads app/etc/local.xml and returns
// DB credentials as a base64-encoded JSON blob (compatible with the Magento 2 probe pattern).
const magento1LocalXMLProbePHP = `
$f='app/etc/local.xml';
if(!file_exists($f)){fwrite(STDERR,"local.xml not found\n");exit(2);}
$x=@simplexml_load_file($f);
if(!$x){fwrite(STDERR,"Failed to parse local.xml\n");exit(3);}
$c=$x->global->resources->default_setup->connection;
if(!$c){fwrite(STDERR,"Connection node missing\n");exit(4);}
$r=[
  "host"    =>(string)$c->host,
  "username"=>(string)$c->username,
  "password"=>(string)$c->password,
  "dbname"  =>(string)$c->dbname,
  "crypt_key"=>(string)($x->global->crypt->key ?? ""),
];
echo base64_encode(json_encode($r));`
