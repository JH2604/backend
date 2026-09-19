package model



type RegisterReq struct {
	Username string `json:"username" binding:"max=20,min=3,required"`
	Password string `json:"password" binding:"max=32,min=6,required"`
}


