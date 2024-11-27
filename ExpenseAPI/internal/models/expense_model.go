package models

type Expense struct {
	ExpenseId   string	`json:"expenseId" bson:"expenseId"`
	UserId      string	`json:"userId" bson:"userId"`
	Description	string	`json:"description" bson:"description"`
	Amount		float32	`json:"amount" bson:"amount"`
	Category	string	`json:"category" bson:"category"`
}