package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ggmolly/belfast/internal/orm"
	"github.com/ggmolly/belfast/internal/protobuf"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func clearCommanderTB(t *testing.T) {
	t.Helper()
	if err := orm.GormDB.Exec("DELETE FROM commander_tbs").Error; err != nil {
		t.Fatalf("clear commander tb: %v", err)
	}
}

func buildTestTBInfoJSON(t *testing.T) string {
	t.Helper()
	info := &protobuf.TBINFO{
		Id: proto.Uint32(1),
		Fsm: &protobuf.TBFSM{
			SystemNo:    proto.Uint32(0),
			CurrentNode: proto.Uint32(0),
			Cache: []*protobuf.TBFSMCACHE{
				{
					CachePlan: []*protobuf.TBFSMCACHEPLAN{{
						CurIndex: proto.Uint32(0),
						Plans:    []*protobuf.KVDATA{},
					}},
					CacheTalent: []*protobuf.TBFSMCACHETALENT{{
						Finished:  proto.Uint32(0),
						Talents:   []uint32{},
						Retalents: []uint32{},
					}},
					CacheSite: []*protobuf.TBFSMCACHESITE{{
						Events:             []uint32{},
						Shops:              []uint32{},
						Buys:               []*protobuf.KVDATA{},
						State:              &protobuf.KVDATA{Key: proto.Uint32(0), Value: proto.Uint32(0)},
						CharacterThisRound: []uint32{},
					}},
					CacheChat: []*protobuf.TBFSMCACHECHAT{{
						Finished: proto.Uint32(0),
						Chats:    []uint32{},
					}},
					CacheEnd: []*protobuf.TBFSMCACHEEND{{
						Ends:   []uint32{},
						Select: proto.Uint32(0),
					}},
					CacheMind: []*protobuf.TBFSMCACHEMIND{{}},
				},
			},
		},
		Round: &protobuf.TBROUND{Round: proto.Uint32(1)},
		Res: &protobuf.TBRES{
			Attrs:    []*protobuf.KVDATA{},
			Resource: []*protobuf.KVDATA{},
		},
		Talent: &protobuf.TBTALENT{Talents: []uint32{}},
		Plan:   &protobuf.TBPLAN{PlanUpgrade: []uint32{}},
		Site: &protobuf.TBSITE{
			Characters:   []uint32{},
			WorkCounter:  []*protobuf.KVDATA{},
			Works:        []uint32{},
			EventCounter: []*protobuf.KVDATA{},
		},
		Evaluations: []*protobuf.KVDATA{},
		Name:        proto.String(""),
		FavorLv:     proto.Uint32(0),
		Benefit: &protobuf.TBBENEFIT{
			Actives:  []*protobuf.TBBF{},
			Pendings: []uint32{},
		},
	}
	data, err := protojson.Marshal(info)
	if err != nil {
		t.Fatalf("marshal tb info: %v", err)
	}
	return string(data)
}

func buildTestTBPermanentJSON(t *testing.T) string {
	t.Helper()
	permanent := &protobuf.TBPERMANENT{
		NgPlusCount:   proto.Uint32(1),
		Polaroids:     []uint32{},
		Endings:       []uint32{},
		ActiveEndings: []uint32{},
	}
	data, err := protojson.Marshal(permanent)
	if err != nil {
		t.Fatalf("marshal tb permanent: %v", err)
	}
	return string(data)
}

func TestPlayerTBNotFound(t *testing.T) {
	app := newPlayerHandlerTestApp(t)
	clearCommanderTB(t)
	clearCommanders(t)
	seedCommander(t, 9101, "TB Tester")
	t.Cleanup(func() {
		clearCommanderTB(t)
		clearCommanders(t)
	})

	request := httptest.NewRequest(http.MethodGet, "/api/v1/players/9101/tb", nil)
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", response.Code)
	}
}

func TestPlayerTBCreateUpdateDeleteFlow(t *testing.T) {
	app := newPlayerHandlerTestApp(t)
	clearCommanderTB(t)
	clearCommanders(t)
	seedCommander(t, 9102, "TB Flow Tester")
	t.Cleanup(func() {
		clearCommanderTB(t)
		clearCommanders(t)
	})

	createBody := `{"tb":` + buildTestTBInfoJSON(t) + `,"permanent":` + buildTestTBPermanentJSON(t) + `}`
	createRequest := httptest.NewRequest(http.MethodPost, "/api/v1/players/9102/tb", strings.NewReader(createBody))
	createRequest.Header.Set("Content-Type", "application/json")
	createResponse := httptest.NewRecorder()
	app.ServeHTTP(createResponse, createRequest)

	if createResponse.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", createResponse.Code, createResponse.Body.String())
	}

	getRequest := httptest.NewRequest(http.MethodGet, "/api/v1/players/9102/tb", nil)
	getResponse := httptest.NewRecorder()
	app.ServeHTTP(getResponse, getRequest)

	if getResponse.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", getResponse.Code)
	}

	var getPayload struct {
		OK bool `json:"ok"`
	}
	if err := json.NewDecoder(getResponse.Body).Decode(&getPayload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !getPayload.OK {
		t.Fatalf("expected ok true")
	}

	updateBody := `{"tb":` + buildTestTBInfoJSON(t) + `,"permanent":` + buildTestTBPermanentJSON(t) + `}`
	updateRequest := httptest.NewRequest(http.MethodPut, "/api/v1/players/9102/tb", strings.NewReader(updateBody))
	updateRequest.Header.Set("Content-Type", "application/json")
	updateResponse := httptest.NewRecorder()
	app.ServeHTTP(updateResponse, updateRequest)

	if updateResponse.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", updateResponse.Code)
	}

	deleteRequest := httptest.NewRequest(http.MethodDelete, "/api/v1/players/9102/tb", nil)
	deleteResponse := httptest.NewRecorder()
	app.ServeHTTP(deleteResponse, deleteRequest)

	if deleteResponse.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", deleteResponse.Code)
	}
}

func TestPlayerTBBadPayloads(t *testing.T) {
	app := newPlayerHandlerTestApp(t)
	clearCommanderTB(t)
	clearCommanders(t)
	seedCommander(t, 9103, "TB Bad Payload")
	t.Cleanup(func() {
		clearCommanderTB(t)
		clearCommanders(t)
	})

	request := httptest.NewRequest(http.MethodPost, "/api/v1/players/9103/tb", strings.NewReader("{invalid"))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.Code)
	}

	request = httptest.NewRequest(http.MethodPost, "/api/v1/players/9103/tb", strings.NewReader(`{"tb":{}}`))
	request.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.Code)
	}

	request = httptest.NewRequest(http.MethodPut, "/api/v1/players/9103/tb", strings.NewReader(`{"tb":{}}`))
	request.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.Code)
	}
}

func TestPlayerTBConflictAndUpdateNotFound(t *testing.T) {
	app := newPlayerHandlerTestApp(t)
	clearCommanderTB(t)
	clearCommanders(t)
	seedCommander(t, 9104, "TB Conflict")
	t.Cleanup(func() {
		clearCommanderTB(t)
		clearCommanders(t)
	})

	createBody := `{"tb":` + buildTestTBInfoJSON(t) + `,"permanent":` + buildTestTBPermanentJSON(t) + `}`
	createRequest := httptest.NewRequest(http.MethodPost, "/api/v1/players/9104/tb", strings.NewReader(createBody))
	createRequest.Header.Set("Content-Type", "application/json")
	createResponse := httptest.NewRecorder()
	app.ServeHTTP(createResponse, createRequest)
	if createResponse.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", createResponse.Code)
	}

	conflictRequest := httptest.NewRequest(http.MethodPost, "/api/v1/players/9104/tb", strings.NewReader(createBody))
	conflictRequest.Header.Set("Content-Type", "application/json")
	conflictResponse := httptest.NewRecorder()
	app.ServeHTTP(conflictResponse, conflictRequest)
	if conflictResponse.Code != http.StatusConflict {
		t.Fatalf("expected status 409, got %d", conflictResponse.Code)
	}

	updateRequest := httptest.NewRequest(http.MethodPut, "/api/v1/players/9105/tb", strings.NewReader(createBody))
	updateRequest.Header.Set("Content-Type", "application/json")
	updateResponse := httptest.NewRecorder()
	app.ServeHTTP(updateResponse, updateRequest)
	if updateResponse.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", updateResponse.Code)
	}
}

func TestPlayerTBDeleteNotFound(t *testing.T) {
	app := newPlayerHandlerTestApp(t)
	clearCommanderTB(t)
	clearCommanders(t)
	seedCommander(t, 9106, "TB Delete")
	t.Cleanup(func() {
		clearCommanderTB(t)
		clearCommanders(t)
	})

	deleteRequest := httptest.NewRequest(http.MethodDelete, "/api/v1/players/9106/tb", nil)
	deleteResponse := httptest.NewRecorder()
	app.ServeHTTP(deleteResponse, deleteRequest)
	if deleteResponse.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", deleteResponse.Code)
	}
}

func TestPlayerTBUpdateInvalidPayload(t *testing.T) {
	app := newPlayerHandlerTestApp(t)
	clearCommanderTB(t)
	clearCommanders(t)
	seedCommander(t, 9107, "TB Update Payload")
	t.Cleanup(func() {
		clearCommanderTB(t)
		clearCommanders(t)
	})

	entry, err := orm.NewCommanderTB(9107, &protobuf.TBINFO{}, &protobuf.TBPERMANENT{})
	if err == nil {
		_ = orm.GormDB.Create(entry).Error
	}

	request := httptest.NewRequest(http.MethodPut, "/api/v1/players/9107/tb", strings.NewReader("{invalid"))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.Code)
	}
}

func TestPlayerTBUpdateInvalidJSONPayload(t *testing.T) {
	app := newPlayerHandlerTestApp(t)
	clearCommanderTB(t)
	clearCommanders(t)
	seedCommander(t, 9108, "TB Invalid JSON")
	t.Cleanup(func() {
		clearCommanderTB(t)
		clearCommanders(t)
	})

	payload := `{"tb":"invalid","permanent":{}}`
	request := httptest.NewRequest(http.MethodPost, "/api/v1/players/9108/tb", strings.NewReader(payload))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	app.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.Code)
	}
}
