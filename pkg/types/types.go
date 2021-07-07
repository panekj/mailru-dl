package types

const (
	APIVersion  int    = 2
	EndpointURL string = "https://cloud.mail.ru/api/v2"
	Build       string = "cloudweb-11674-72-8-0.202012151755"
)

type Response struct {
	Body   Body  `json:"body"`
	Time   int64 `json:"time"`
	Status int   `json:"status"`
}

type Count struct {
	Folders int `json:"folders"`
	Files   int `json:"files"`
}

type Sort struct {
	Order string `json:"order"`
	Type  string `json:"type"`
}

type List struct {
	Count   Count  `json:"count"`
	Name    string `json:"name"`
	Grev    int    `json:"grev"`
	Size    int64  `json:"size"`
	Kind    string `json:"kind"`
	Weblink string `json:"weblink"`
	Type    string `json:"type"`
}

type Body struct {
	WeblinkAccessRights string              `json:"weblink_access_rights"`
	Count               Count               `json:"count"`
	Name                string              `json:"name"`
	Grev                int                 `json:"grev"`
	Size                int64               `json:"size"`
	Sort                Sort                `json:"sort"`
	Kind                string              `json:"kind"`
	Weblink             string              `json:"weblink"`
	Type                string              `json:"type"`
	List                []List              `json:"list"`
	Video               []Video             `json:"video"`
	ViewDirect          []ViewDirect        `json:"view_direct"`
	WeblinkView         []WeblinkView       `json:"weblink_view"`
	WeblinkVideo        []WeblinkVideo      `json:"weblink_video"`
	WeblinkGet          []WeblinkGet        `json:"weblink_get"`
	Stock               []Stock             `json:"stock"`
	WeblinkThumbnails   []WeblinkThumbnails `json:"weblink_thumbnails"`
	PublicUpload        []PublicUpload      `json:"public_upload"`
	Auth                []Auth              `json:"auth"`
	Web                 []Web               `json:"web"`
	View                []View              `json:"view"`
	Upload              []Upload            `json:"upload"`
	Get                 []Get               `json:"get"`
	Thumbnails          []Thumbnails        `json:"thumbnails"`
}

type Video struct {
	Count string `json:"count"`
	URL   string `json:"url"`
}

type ViewDirect struct {
	Count string `json:"count"`
	URL   string `json:"url"`
}

type WeblinkView struct {
	Count string `json:"count"`
	URL   string `json:"url"`
}

type WeblinkVideo struct {
	Count string `json:"count"`
	URL   string `json:"url"`
}

type WeblinkGet struct {
	Count int    `json:"count"`
	URL   string `json:"url"`
}

type Stock struct {
	Count string `json:"count"`
	URL   string `json:"url"`
}

type WeblinkThumbnails struct {
	Count string `json:"count"`
	URL   string `json:"url"`
}

type PublicUpload struct {
	Count string `json:"count"`
	URL   string `json:"url"`
}

type Auth struct {
	Count string `json:"count"`
	URL   string `json:"url"`
}

type Web struct {
	Count string `json:"count"`
	URL   string `json:"url"`
}

type View struct {
	Count string `json:"count"`
	URL   string `json:"url"`
}

type Upload struct {
	Count string `json:"count"`
	URL   string `json:"url"`
}

type Get struct {
	Count string `json:"count"`
	URL   string `json:"url"`
}

type Thumbnails struct {
	Count string `json:"count"`
	URL   string `json:"url"`
}
