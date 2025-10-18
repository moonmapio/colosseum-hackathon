package publisher

// 7) Seguridad / permisos
// En presign, valida que userId puede escribir en ese scopeType/scopeId:
// USER: scopeId == userId
// COMMUNITY: moderador/autor
// PROJECT: owner/admin del proyecto
// En lectura, puedes exponer GET por scope o por entityId.

func HasPermissions(userId, scopeType, scopeId string) bool {
	return true
}
