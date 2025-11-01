package dbHandlers

import(
	"github.com/Irlandesee/GO-Fuels/src/utils/dbHandlers/db"
)

type Handler interface {
	Postgres *db.PostgresHandler
	Mongo   *db.MongoHandler
	Rabbit *rabbit.RabbitHandler
}

func NewDbHandler() *Handler {
	return &Handler{}
}