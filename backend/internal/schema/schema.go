package schema

type SignUp struct {
	Name     string `json:"name" binding:"required"`
	Password string `json:"password" binding:"required"`
	Access   int    `json:"access" binding:"required"`
}

type SignIn struct {
	Name     string `json:"name" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type Publish struct {
	Id     int    `json:"id"`
	Text   string `json:"text" binding:"required"`
	Access int    `json:"access" binding:"required"`
	Name   string `json:"name" binding:"required"`
	UserId int    `json:"user_id" binding:"required"`
}

type AllArticles struct {
	Articles []Publish `json:"articles"`
}
