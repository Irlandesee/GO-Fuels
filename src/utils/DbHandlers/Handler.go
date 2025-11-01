package DbHandlers

import(
	"github.com/Irlandesee/GO-Fuels/src/util/DbHandlers/db"
)

type Handler interface {
	Postgres *db.PostgresHandler
	Mongo   *db.MongoHandler
	Rabbit *rabbit.RabbitHandler
}

func NewDbHandler() *Handler {
	return &Handler{}
}