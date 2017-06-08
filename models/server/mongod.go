package server


import (
	yaml "gopkg.in/yaml.v2"
	"os"
	"path/filepath"
	"github.com/ruslanBik4/httpgo/models/logs"
)

type mongodConfig struct {
	systemPath  string
	wwwPath     string
	SessionPath string
	dbParams struct {
			    DB   string `yaml:"dbName"`
		    }
}

var mConfig *mongodConfig

func GetMongodConfig() *mongodConfig {

	if mConfig != nil {
		return mConfig
	} else {
		mConfig = &mongodConfig{}
	}

	return mConfig
}
func (mConfig *mongodConfig) Init(f_static, f_web, f_session *string) error{

	mConfig.systemPath = *f_static
	mConfig.wwwPath     = *f_web
	mConfig.SessionPath = *f_session

	f, err := os.Open(filepath.Join(mConfig.systemPath, "config/mongo.yml" ))
	if err != nil {
		return err
	}
	fileInfo, _ := f.Stat()
	b  := make([]byte, fileInfo.Size())
	if n, err := f.Read(b); err != nil {
		logs.ErrorLog(err, "n=", n)
		return err

	}

	if err := yaml.Unmarshal(b, &mConfig.dbParams); err != nil {
		return err
	}

	return nil
}

//The Data Source DB has a common format, like e.g. PEAR DB uses it,
// but without type-prefix (optional parts marked by squared brackets):
//
//[username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]
func (mConfig *mongodConfig) MongoDBName() string {
	return mConfig.dbParams.DB
}
