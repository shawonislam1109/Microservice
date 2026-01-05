package model

type Role string

const (
	SuperAdmin  Role = "super_admin"
	Admin       Role = "admin"
	Merchant    Role = "merchant"
	Reseller    Role = "reseller"
	SubReseller Role = "sub_reseller"
	Employee    Role = "employee"
	Client      Role = "client"
)
