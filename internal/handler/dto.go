package handler

type SaveOriginRequest struct {
	URL string `json:"url"`
}
type SaveOriginResponse struct {
	Short string `json:"short_url"`
}

type GetOriginRequest struct {
	Short string `json:"short_url"`
}

type GetOriginResponse struct {
	URL string `json:"url"`
}
