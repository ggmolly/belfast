package types

import "github.com/ggmolly/belfast/internal/orm"

type PlayerLoveLetterStateResponse struct {
	Medals         []orm.LoveLetterMedalState    `json:"medals"`
	ManualLetters  []orm.LoveLetterLetterState   `json:"manual_letters"`
	ConvertedItems []orm.LoveLetterConvertedItem `json:"converted_items"`
	RewardedIDs    []uint32                      `json:"rewarded_ids"`
	LetterContents map[uint32]string             `json:"letter_contents"`
}

type PlayerLoveLetterStateUpdateRequest struct {
	Medals         *[]orm.LoveLetterMedalState    `json:"medals"`
	ManualLetters  *[]orm.LoveLetterLetterState   `json:"manual_letters"`
	ConvertedItems *[]orm.LoveLetterConvertedItem `json:"converted_items"`
	RewardedIDs    *[]uint32                      `json:"rewarded_ids"`
	LetterContents *map[uint32]string             `json:"letter_contents"`
}
