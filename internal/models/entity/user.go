package entity

type User struct {
	ID       string `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name     string `json:"name" gorm:"column:name"`
	Email    string `json:"email" gorm:"column:email"`
	Password string `json:"password" gorm:"column:password"`
	Role     string `json:"role" gorm:"column:role"`
	Lead     []Lead `gorm:"foreignKey:AssignedTo"`
}
