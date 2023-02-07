package data

/*
Data for local testing purposes
*/

type Chat struct {
	IsGroupChat bool    `json:"isGroupChat"`
	Users       []*User `json:"users"`
	Id          string  `json:"_id"`
	ChatName    string  `json:"chatName"`
}

type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Chats []*Chat

var ChatsOfUsers = []*Chat{
	{
		IsGroupChat: false,
		Users: []*User{
			{
				Name:  "John Mark",
				Email: "johnmark@gmail.com",
			},
			{
				Name:  "Jack Sparrow",
				Email: "jacksparrow@gmail.com",
			},
		},
		Id:       "617a077e18c25468bc7c4dd4",
		ChatName: "John mark",
	},
}
