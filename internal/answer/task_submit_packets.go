package answer

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/ggmolly/belfast/internal/connection"
	"github.com/ggmolly/belfast/internal/db"
	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/proto"
)

const (
	taskDataTemplateCategory = "ShareCfg/task_data_template.json"

	taskResultSuccess = uint32(0)
	taskResultFailure = uint32(1)

	taskFinishedFlagBase  = uint32(2_000_000)
	taskSubmittedFlagBase = uint32(3_000_000)
)

type taskTemplate struct {
	ID           uint32          `json:"id"`
	TargetNum    json.RawMessage `json:"target_num"`
	AwardDisplay json.RawMessage `json:"award_display"`
	AwardChoice  json.RawMessage `json:"award_choice"`
}

type taskFlagState struct {
	commanderID uint32
	flags       map[uint32]struct{}
}

func TaskUpdateProgress(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_20009
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 20010, err
	}

	state, err := loadTaskFlagState(client.Commander.CommanderID)
	if err != nil {
		return 0, 20010, err
	}

	response := protobuf.SC_20010{Result: proto.Uint32(taskResultSuccess)}
	for _, update := range payload.GetProgressinfo() {
		taskID := update.GetId()
		if taskID == 0 {
			response.Result = proto.Uint32(taskResultFailure)
			break
		}
		template, ok, err := loadTaskTemplate(taskID)
		if err != nil {
			return 0, 20010, err
		}
		if !ok {
			response.Result = proto.Uint32(taskResultFailure)
			break
		}
		targetNum := parseTaskTargetNum(template.TargetNum)
		if update.GetProgress() >= targetNum {
			if err := state.set(taskFinishedFlagID(taskID)); err != nil {
				return 0, 20010, err
			}
			if err := state.clear(taskSubmittedFlagID(taskID)); err != nil {
				return 0, 20010, err
			}
		}
	}

	return client.SendMessage(20010, &response)
}

func TaskSubmitSingle(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_20005
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 20006, err
	}

	state, err := loadTaskFlagState(client.Commander.CommanderID)
	if err != nil {
		return 0, 20006, err
	}

	response := protobuf.SC_20006{Result: proto.Uint32(taskResultFailure), AwardList: []*protobuf.DROPINFO{}}
	drops, ok, err := submitTask(client, state, payload.GetId(), payload.GetChoiceAward(), true)
	if err != nil {
		return 0, 20006, err
	}
	if ok {
		response.Result = proto.Uint32(taskResultSuccess)
		response.AwardList = drops
	}

	return client.SendMessage(20006, &response)
}

func TaskSubmitOneStep(buffer *[]byte, client *connection.Client) (int, int, error) {
	var payload protobuf.CS_20011
	if err := proto.Unmarshal(*buffer, &payload); err != nil {
		return 0, 20012, err
	}

	state, err := loadTaskFlagState(client.Commander.CommanderID)
	if err != nil {
		return 0, 20012, err
	}

	response := protobuf.SC_20012{IdList: []uint32{}, AwardList: []*protobuf.DROPINFO{}}
	mergedDrops := make(map[string]*protobuf.DROPINFO)
	for _, taskID := range payload.GetIdList() {
		drops, ok, err := submitTask(client, state, taskID, nil, false)
		if err != nil {
			return 0, 20012, err
		}
		if !ok {
			continue
		}
		response.IdList = append(response.IdList, taskID)
		for _, drop := range drops {
			key := fmt.Sprintf("%d_%d", drop.GetType(), drop.GetId())
			existing := mergedDrops[key]
			if existing == nil {
				mergedDrops[key] = newDropInfo(drop.GetType(), drop.GetId(), drop.GetNumber())
				continue
			}
			existing.Number = proto.Uint32(existing.GetNumber() + drop.GetNumber())
		}
	}
	response.AwardList = dropMapToList(mergedDrops)

	return client.SendMessage(20012, &response)
}

func submitTask(client *connection.Client, state *taskFlagState, taskID uint32, choiceAward []*protobuf.DROPINFO, allowChoice bool) ([]*protobuf.DROPINFO, bool, error) {
	if taskID == 0 {
		return nil, false, nil
	}
	template, ok, err := loadTaskTemplate(taskID)
	if err != nil {
		return nil, false, err
	}
	if !ok {
		return nil, false, nil
	}
	if !state.has(taskFinishedFlagID(taskID)) || state.has(taskSubmittedFlagID(taskID)) {
		return nil, false, nil
	}

	rewardDrops, ok := resolveTaskRewardDrops(template, choiceAward, allowChoice)
	if !ok {
		return nil, false, nil
	}

	if err := ensureCommanderLoaded(client, "TaskSubmit"); err != nil {
		return nil, false, err
	}
	if err := applyDropList(client, rewardDrops); err != nil {
		return nil, false, err
	}

	if err := state.set(taskSubmittedFlagID(taskID)); err != nil {
		return nil, false, err
	}
	if err := state.clear(taskFinishedFlagID(taskID)); err != nil {
		return nil, false, err
	}

	return dropMapToList(rewardDrops), true, nil
}

func resolveTaskRewardDrops(template taskTemplate, choiceAward []*protobuf.DROPINFO, allowChoice bool) (map[string]*protobuf.DROPINFO, bool) {
	choiceOptions, ok := parseTaskRewardOptions(template.AwardChoice)
	if !ok {
		return nil, false
	}
	var selected [][]uint32
	if len(choiceOptions) > 0 {
		if !allowChoice {
			return nil, false
		}
		selected, ok = selectTaskChoice(choiceOptions, choiceAward)
		if !ok {
			return nil, false
		}
	} else {
		if len(choiceAward) > 0 {
			return nil, false
		}
		selected, ok = parseTaskDropList(template.AwardDisplay)
		if !ok {
			return nil, false
		}
	}

	drops := make(map[string]*protobuf.DROPINFO)
	for _, entry := range selected {
		if len(entry) < 3 {
			return nil, false
		}
		dropType, dropID, amount := entry[0], entry[1], entry[2]
		key := fmt.Sprintf("%d_%d", dropType, dropID)
		existing := drops[key]
		if existing == nil {
			drops[key] = newDropInfo(dropType, dropID, amount)
			continue
		}
		existing.Number = proto.Uint32(existing.GetNumber() + amount)
	}

	return drops, true
}

func selectTaskChoice(options [][][]uint32, choiceAward []*protobuf.DROPINFO) ([][]uint32, bool) {
	if len(choiceAward) != 1 {
		return nil, false
	}
	selected := []uint32{choiceAward[0].GetType(), choiceAward[0].GetId(), choiceAward[0].GetNumber()}
	for _, option := range options {
		if len(option) == 0 {
			continue
		}
		if option[0][0] == selected[0] && option[0][1] == selected[1] && option[0][2] == selected[2] {
			return option, true
		}
	}
	return nil, false
}

func parseTaskRewardOptions(raw json.RawMessage) ([][][]uint32, bool) {
	if len(raw) == 0 {
		return nil, true
	}
	var payload any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, false
	}
	items, ok := payload.([]any)
	if !ok || len(items) == 0 {
		return nil, true
	}

	firstList, firstIsList := items[0].([]any)
	if !firstIsList || len(firstList) == 0 {
		return nil, false
	}

	if _, isNumber := parseJSONUint(firstList[0]); isNumber {
		drops, ok := parseTaskDropList(raw)
		if !ok {
			return nil, false
		}
		return [][][]uint32{drops}, true
	}

	options := make([][][]uint32, 0, len(items))
	for _, optionRaw := range items {
		optionJSON, err := json.Marshal(optionRaw)
		if err != nil {
			return nil, false
		}
		drops, ok := parseTaskDropList(optionJSON)
		if !ok {
			return nil, false
		}
		options = append(options, drops)
	}
	return options, true
}

func parseTaskDropList(raw json.RawMessage) ([][]uint32, bool) {
	if len(raw) == 0 {
		return nil, true
	}
	var payload []any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, false
	}
	drops := make([][]uint32, 0, len(payload))
	for _, entry := range payload {
		parts, ok := entry.([]any)
		if !ok || len(parts) < 3 {
			return nil, false
		}
		dropType, ok := parseJSONUint(parts[0])
		if !ok {
			return nil, false
		}
		dropID, ok := parseJSONUint(parts[1])
		if !ok {
			return nil, false
		}
		amount, ok := parseJSONUint(parts[2])
		if !ok {
			return nil, false
		}
		drops = append(drops, []uint32{dropType, dropID, amount})
	}
	return drops, true
}

func loadTaskTemplate(taskID uint32) (taskTemplate, bool, error) {
	entry, err := orm.GetConfigEntry(taskDataTemplateCategory, strconv.FormatUint(uint64(taskID), 10))
	if err != nil {
		if db.IsNotFound(err) {
			return taskTemplate{}, false, nil
		}
		return taskTemplate{}, false, err
	}
	var template taskTemplate
	if err := json.Unmarshal(entry.Data, &template); err != nil {
		return taskTemplate{}, false, err
	}
	if template.ID == 0 {
		template.ID = taskID
	}
	return template, true, nil
}

func parseTaskTargetNum(raw json.RawMessage) uint32 {
	if len(raw) == 0 {
		return 1
	}
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return 1
	}
	if number, ok := parseJSONUint(value); ok {
		if number == 0 {
			return 1
		}
		return number
	}
	values, ok := value.([]any)
	if ok && len(values) > 0 {
		if number, ok := parseJSONUint(values[0]); ok {
			if number == 0 {
				return 1
			}
			return number
		}
	}
	return 1
}

func loadTaskFlagState(commanderID uint32) (*taskFlagState, error) {
	flags, err := orm.ListCommanderCommonFlags(commanderID)
	if err != nil {
		return nil, err
	}
	flagMap := make(map[uint32]struct{}, len(flags))
	for _, flagID := range flags {
		flagMap[flagID] = struct{}{}
	}
	return &taskFlagState{commanderID: commanderID, flags: flagMap}, nil
}

func (s *taskFlagState) has(flagID uint32) bool {
	_, ok := s.flags[flagID]
	return ok
}

func (s *taskFlagState) set(flagID uint32) error {
	if s.has(flagID) {
		return nil
	}
	if err := orm.SetCommanderCommonFlag(s.commanderID, flagID); err != nil {
		return err
	}
	s.flags[flagID] = struct{}{}
	return nil
}

func (s *taskFlagState) clear(flagID uint32) error {
	if !s.has(flagID) {
		return nil
	}
	if err := orm.ClearCommanderCommonFlag(s.commanderID, flagID); err != nil {
		return err
	}
	delete(s.flags, flagID)
	return nil
}

func taskFinishedFlagID(taskID uint32) uint32 {
	return taskFinishedFlagBase + taskID
}

func taskSubmittedFlagID(taskID uint32) uint32 {
	return taskSubmittedFlagBase + taskID
}
