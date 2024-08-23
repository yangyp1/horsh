package v1

var (
	// common errors
	ErrSuccess             = newError(0, "ok")
	ErrBadRequest          = newError(400, "Bad Request")
	ErrUnauthorized        = newError(401, "Unauthorized")
	ErrNotFound            = newError(404, "Not Found")
	ErrInternalServerError = newError(500, "Internal Server Error")

	// more biz errors
	ErrAddressAlreadyExist = newError(1001, "The Address is already Exist.")
	ErrInvalidInviteCode        = newError(1002, "Invalid InviteCode")
	ErrUserNotExist        = newError(1003, "The User is Not Exist.")
)
