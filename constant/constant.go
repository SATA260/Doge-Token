package constant

import (
	"math"
	"time"
)

const (
	LOGIN_URL  = "/login"
	LOGOUT_URL = "/logout"

	DOGE_TOKEN             = "doge_token"
	DOGE_USER              = "doge_user"
	DOGE_USER_STORE_PREFIX = "doge_user:"

	CLIENT_REDIRECT_URL = "/login"

	CODE_LOGIN_FAILED = "501"

	SECRET_KEY = "doge_token_666"

	DEFAULT_EXPIRE_TIME = EXPIRE_TIME_1_DAY

	EXPIRE_TIME_1_DAY       = int64(time.Hour) * 24
	EXPIRE_TIME_FOR_7_DAY   = EXPIRE_TIME_1_DAY * 7
	EXPIRE_TIME_FOR_30_DAY  = EXPIRE_TIME_1_DAY * 30
	EXPIRE_TIME_FOR_1_YEAR  = EXPIRE_TIME_1_DAY * 365
	EXPIRE_TIME_FOR_10_YEAR = EXPIRE_TIME_1_DAY * 365 * 10

	COOKIE_MAX_AGE = math.MaxInt64
	COOKIE_PATH    = "/"
)
