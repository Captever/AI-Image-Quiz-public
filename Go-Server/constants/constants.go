package constants

// 퀴즈 이미지 기본 양식 관련
const (
	IMAGE_COUNT_ROW    = 3
	IMAGE_COUNT_COLUMN = 3
	IMAGE_COUNT        = IMAGE_COUNT_ROW * IMAGE_COUNT_COLUMN
	IMAGE_SIZE         = 512
	SPRITE_SHEET_SIZE  = 1536 // 3 * 512
)

// 게임 진행 관련
const (
	CATEGORY_COUNT = 3

	MAX_QUIZ_COUNT       = 9 // 게임 당 최대 퀴즈 이미지 개수
	CORRECTED_QUIZ_COUNT = 5 // 게임 당 최대 완전히 맞춘 퀴즈 이미지 개수

	MAX_SCORE_PER_ANSWER = 5
	START_COUNTDOWN_TIME = 5
	QUIZ_COUNTDOWN_TIME  = 5
	QUIZ_TIMER_TIME      = 30
)
