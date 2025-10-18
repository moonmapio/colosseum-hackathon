package constants

const (
	HeaderHost            = "host"
	HeaderUserAgent       = "user-agent"
	HeaderCFRay           = "cf-ray"
	HeaderCFIPCountry     = "cf-ipcountry"
	HeaderCFConnectingIP  = "cf-connecting-ip"
	HeaderXForwardedFor   = "x-forwarded-for"
	HeaderXRealIP         = "x-real-ip"
	HeaderAcceptLanguage  = "accept-language"
	HeaderReferer         = "referer"
	HeaderOrigin          = "origin"
	HeaderSecCHUA         = "sec-ch-ua"
	HeaderSecCHUAPlatform = "sec-ch-ua-platform"
	HeaderWSVersion       = "sec-websocket-version"
	HeaderWSExtensions    = "sec-websocket-extensions"
	HeaderWSSubprotocol   = "sec-websocket-protocol"
)

var HeaderList = []string{
	HeaderHost,
	HeaderUserAgent,
	HeaderCFRay,
	HeaderCFIPCountry,
	HeaderCFConnectingIP,
	HeaderXForwardedFor,
	HeaderXRealIP,
	HeaderAcceptLanguage,
	HeaderReferer,
	HeaderOrigin,
	HeaderSecCHUA,
	HeaderSecCHUAPlatform,
	HeaderWSVersion,
	HeaderWSExtensions,
	HeaderWSSubprotocol,
}
