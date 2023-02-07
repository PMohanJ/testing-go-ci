package controllers

import (
	"go.mongodb.org/mongo-driver/bson"
)

func MatchStageBySingleField(key string, value interface{}) bson.D {
	return bson.D{
		{
			"$match", bson.D{
				{key, value},
			},
		},
	}
}

func LookUpStage(from, localField, foreignField, as string) bson.D {
	return bson.D{
		{
			"$lookup", bson.D{
				{"from", from},
				{"localField", localField},
				{"foreignField", foreignField},
				{"as", as},
			},
		},
	}
}

func ProjectStage(field1, field2, field3, field4, field5 string) bson.D {
	return bson.D{
		{
			"$project", bson.D{
				{field1, 0},
				{field2, 0},
				{field3, 0},
				{field4, 0},
				{field5, 0},
			},
		},
	}
}
