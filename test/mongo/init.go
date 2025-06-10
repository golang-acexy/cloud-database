package mongo

import (
	"github.com/golang-acexy/starter-mongo/mongostarter"
	"github.com/golang-acexy/starter-parent/parent"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var loader *parent.StarterLoader

func init() {
	loader = parent.NewStarterLoader([]parent.Starter{
		&mongostarter.MongoStarter{
			Config: mongostarter.MongoConfig{
				MongoUri: "mongodb://acexy:tech-acexy@localhost:27017/local?authSource=admin",
				//Database: "local",
				BsonOpts: &options.BSONOptions{
					UseJSONStructTags:   true,
					ObjectIDAsHexString: true,
					OmitZeroStruct:      true,
					ZeroStructs:         true,
				},
				EnableLogger: true,
			},
		},
	})
	err := loader.Start()
	if err != nil {
		println(err)
		return
	}
}
