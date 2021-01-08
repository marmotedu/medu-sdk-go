package request

type Request interface {
	GetURL() string
	GetMethod() string
	GetVersion() string
	GetHeaders() map[string]string
}

// Request is the base struct of service requests
type BaseRequest struct {
	URL     string // resource url, i.e. /regions/${regionId}/elasticIps/${elasticIpId}
	Method  string
	Header  map[string]string
	Version string
}

func (r BaseRequest) GetURL() string {
	return r.URL
}

func (r BaseRequest) GetMethod() string {
	return r.Method
}

func (r BaseRequest) GetVersion() string {
	return r.Version
}

func (r BaseRequest) GetHeaders() map[string]string {
	return r.Header
}

// AddHeader only adds pin or erp, they will be encoded to base64 code
func (r *BaseRequest) AddHeader(key, value string) {
	if r.Header == nil {
		r.Header = make(map[string]string)
	}
	r.Header[key] = value
}
