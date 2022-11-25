package main


type UserDao interface {
	FindUserById()
}

type userDao struct{

}

func (u userDao) FindUserById() {

}

func NewUserDao() UserDao {
  return userDao{}
}