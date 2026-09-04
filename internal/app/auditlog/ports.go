package auditlog

import (
	"context"

	domainauditlog "github.com/ESG-Project/suassu-api/internal/domain/auditlog"
)

// Repo persiste entradas da trilha de auditoria.
type Repo interface {
	Create(ctx context.Context, l *domainauditlog.Log) error
}

// Recorder é o que os demais serviços consomem para registrar um evento.
//
// Não devolve erro de propósito: uma falha ao gravar a auditoria não pode
// transformar uma operação já concluída em erro para o operador. O user-crud
// faz o contrário (o await do CreateLogService derruba o request com 400
// depois de o registro já ter sido criado), o que leva o usuário a repetir a
// ação e duplicar dados.
type Recorder interface {
	Record(ctx context.Context, actorID *string, enterpriseID, description string)
}
