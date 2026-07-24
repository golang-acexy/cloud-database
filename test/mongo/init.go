package mongo

import (
	"github.com/golang-acexy/starter-mongo/mongostarter"
	"github.com/golang-acexy/starter-parent/parent"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var loader *parent.StarterLoader

func init() {
	loader = parent.InitStarterLoader([]parent.Starter{
		&mongostarter.MongoStarter{
			Config: mongostarter.MongoConfig{
				MongoURI: "mongodb://acexy:tech-acexy@localhost:27017/local?authSource=admin",
				//Database: "local",
				BSONOptions: &options.BSONOptions{
					UseJSONStructTags:   true,
					ObjectIDAsHexString: true,
					OmitZeroStruct:      true,
					ZeroStructs:         true,
				},
				EnableLogger: true,
			},
		},
	})
}
