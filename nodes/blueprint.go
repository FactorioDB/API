package nodes

import (
	"net/http"

	"time"

	"strconv"

	"github.com/BlooperDB/API/api"
	"github.com/BlooperDB/API/db"
	"github.com/BlooperDB/API/storage"
	"github.com/BlooperDB/API/utils"
	"github.com/gorilla/mux"
)

type BlueprintResponse struct {
	Id          uint        `json:"id"`
	UserId      uint        `json:"user"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Latest      uint        `json:"latest-revision"`
	Revisions   []*Revision `json:"revisions,omitempty"`
	Tags        []string    `json:"tags"`
	CreatedAt   time.Time   `json:"created-at"`
	UpdatedAt   time.Time   `json:"updated-at"`
}

func RegisterBlueprintRoutes(router api.RegisterRoute) {
	router("GET", "/blueprints", getBlueprints)
	router("GET", "/blueprints/popular", popularBlueprints)
	router("GET", "/blueprints/search/", searchBlueprints)
	router("GET", "/blueprints/search/{query}", searchBlueprints)

	router("POST", "/blueprint", api.AuthHandler(postBlueprint, true))
	router("GET", "/blueprint/{blueprint}", getBlueprint)
	router("PUT", "/blueprint/{blueprint}", api.AuthHandler(updateBlueprint, true))
	router("DELETE", "/blueprint/{blueprint}", api.AuthHandler(deleteBlueprint, true))

	router("GET", "/blueprint/{blueprint}/revisions", getRevisions)
	router("GET", "/blueprint/{blueprint}/revision/latest", getRevisionLatest)
	router("GET", "/blueprint/{blueprint}/revision/{revision}", getRevisionIncremental)
}

type SearchBlueprintsResponse struct {
	Blueprints []*BlueprintResponse `json:"blueprints"`
}

/*
Search for blueprints
*/
func searchBlueprints(r *http.Request) (interface{}, *utils.ErrorResponse) {
	var (
		offset, _ = strconv.Atoi(r.URL.Query().Get("offset"))
		count, _  = strconv.Atoi(r.URL.Query().Get("count"))
	)

	if count == 0 {
		count = 20
	}
	if count > 100 {
		count = 100
	}

	query, ok := mux.Vars(r)["query"]
	if !ok {
		return nil, &utils.Error_no_search_terms
	}

	blueprints := db.SearchBlueprints(query, offset, count)
	reBlueprint := make([]*BlueprintResponse, len(blueprints))

	for i, blueprint := range blueprints {
		var revId uint = 0
		if rev := blueprint.GetLatestRevision(); rev != nil {
			revId = rev.Revision
		}

		tags := blueprint.GetTags()
		reTags := make([]string, len(tags))

		for i, tag := range tags {
			reTags[i] = tag.Name
		}

		reBlueprint[i] = &BlueprintResponse{
			Id:          blueprint.ID,
			UserId:      blueprint.UserID,
			Name:        blueprint.Name,
			Description: blueprint.Description,
			CreatedAt:   blueprint.CreatedAt,
			UpdatedAt:   blueprint.UpdatedAt,
			Latest:      revId,
			Tags:        reTags,
		}
	}

	return SearchBlueprintsResponse{
		Blueprints: reBlueprint,
	}, nil
}

/*
Get popular blueprints
*/
func popularBlueprints(r *http.Request) (interface{}, *utils.ErrorResponse) {
	var (
		offset, _ = strconv.Atoi(r.URL.Query().Get("offset"))
		count, _  = strconv.Atoi(r.URL.Query().Get("count"))
	)

	if count == 0 {
		count = 20
	}
	if count > 100 {
		count = 100
	}

	blueprints := db.PopularBlueprints(offset, count)
	reBlueprint := make([]*BlueprintResponse, len(blueprints))

	for i, blueprint := range blueprints {
		var revId uint = 0
		if rev := blueprint.GetLatestRevision(); rev != nil {
			revId = rev.Revision
		}

		tags := blueprint.GetTags()
		reTags := make([]string, len(tags))

		for i, tag := range tags {
			reTags[i] = tag.Name
		}

		reBlueprint[i] = &BlueprintResponse{
			Id:          blueprint.ID,
			UserId:      blueprint.UserID,
			Name:        blueprint.Name,
			Description: blueprint.Description,
			CreatedAt:   blueprint.CreatedAt,
			UpdatedAt:   blueprint.UpdatedAt,
			Latest:      revId,
			Tags:        reTags,
		}
	}

	return SearchBlueprintsResponse{
		Blueprints: reBlueprint,
	}, nil
}

type GetBlueprintsResponse struct {
	Blueprints []*BlueprintResponse `json:"blueprints"`
}

/*
Get all blueprints (paged)
*/
func getBlueprints(r *http.Request) (interface{}, *utils.ErrorResponse) {
	var (
		offset, _ = strconv.Atoi(r.URL.Query().Get("offset"))
		count, _  = strconv.Atoi(r.URL.Query().Get("count"))
	)

	if count == 0 {
		count = 20
	}
	if count > 100 {
		count = 100
	}

	blueprints := db.GetAllBlueprints(offset, count)
	reBlueprint := make([]*BlueprintResponse, len(blueprints))

	for i, blueprint := range blueprints {
		var revId uint = 0
		if rev := blueprint.GetLatestRevision(); rev != nil {
			revId = rev.Revision
		}

		tags := blueprint.GetTags()
		reTags := make([]string, len(tags))

		for i, tag := range tags {
			reTags[i] = tag.Name
		}

		reBlueprint[i] = &BlueprintResponse{
			Id:          blueprint.ID,
			UserId:      blueprint.UserID,
			Name:        blueprint.Name,
			Description: blueprint.Description,
			CreatedAt:   blueprint.CreatedAt,
			UpdatedAt:   blueprint.UpdatedAt,
			Latest:      revId,
			Tags:        reTags,
		}
	}

	return GetBlueprintsResponse{
		Blueprints: reBlueprint,
	}, nil
}

/*
Get a specific blueprint
*/
func getBlueprint(r *http.Request) (interface{}, *utils.ErrorResponse) {
	blueprint, e := parseBlueprint(r)

	if e != nil {
		return nil, e
	}

	getRevisions := len(r.URL.Query()["revisions"]) > 0
	getComments := len(r.URL.Query()["comments"]) > 0

	var reRevision []*Revision

	if getRevisions {
		revisions := blueprint.GetRevisions()
		reRevision = make([]*Revision, len(revisions))

		authUser := db.GetAuthUser(r)

		for i, revision := range revisions {
			if revision.DeletedAt != nil {
				continue
			}
			rev, err := revisionToJSON(authUser, &revision, getComments)
			if err != nil {
				return nil, err
			}
			reRevision[i] = rev
		}
	}

	tags := blueprint.GetTags()
	reTags := make([]string, len(tags))

	for i, tag := range tags {
		reTags[i] = tag.Name
	}

	var revId uint = 0
	if rev := blueprint.GetLatestRevision(); rev != nil {
		revId = rev.Revision
	}

	return BlueprintResponse{
		Id:          blueprint.ID,
		UserId:      blueprint.UserID,
		Name:        blueprint.Name,
		Description: blueprint.Description,
		CreatedAt:   blueprint.CreatedAt,
		UpdatedAt:   blueprint.UpdatedAt,
		Latest:      revId,
		Revisions:   reRevision,
		Tags:        reTags,
	}, nil
}

type PostBlueprintRequest struct {
	Name            string   `json:"name" validate:"min=5" `
	Description     string   `json:"description" validate:"nonzero" `
	BlueprintString string   `json:"blueprint-string" validate:"nonzero,blueprint_string"`
	Tags            []string `json:"tags" validate:"min=1"`
}

type PostBlueprintResponse struct {
	BlueprintId uint `json:"blueprint-id"`

	// Global unique revision identifier
	RevisionId uint `json:"revision-id"`

	// Blueprint incremental version
	Revision uint `json:"revision"`
}

/*
Post a new blueprint
*/
func postBlueprint(u *db.User, r *http.Request) (interface{}, *utils.ErrorResponse) {
	var request PostBlueprintRequest
	e := utils.ValidateRequestBody(r, &request)

	if e != nil {
		return nil, e
	}

	sha265 := utils.SHA265(request.BlueprintString)

	if db.FindRevisionByChecksum(sha265) != nil {
		return nil, &utils.Error_blueprint_string_already_exists
	}

	blueprint := &db.Blueprint{
		UserID:       u.ID,
		Name:         request.Name,
		Description:  request.Description,
		LastRevision: 1,
	}

	blueprint.Save()

	bpVersion, _ := strconv.Atoi(request.BlueprintString[0:1])

	revision := &db.Revision{
		BlueprintID:       blueprint.ID,
		Revision:          blueprint.LastRevision,
		Changes:           "",
		BlueprintVersion:  bpVersion,
		BlueprintChecksum: sha265,
	}

	revision.Save()

	storage.SaveRevision(revision.ID, request.BlueprintString)

	for _, tag := range request.Tags {

		t := db.GetTagByName(tag)

		if t == nil {
			t = &db.Tag{
				Name: tag,
			}

			t.Save()
		}

		bt := db.BlueprintTag{
			BlueprintId: blueprint.ID,
			TagId:       t.ID,
		}

		bt.Save()
	}

	return PostBlueprintResponse{
		BlueprintId: blueprint.ID,
		RevisionId:  revision.ID,
		Revision:    revision.Revision,
	}, nil
}

type PutBlueprintRequest struct {
	Name        string   `json:"name" validate:"min=5"`
	Description string   `json:"description" validate:"nonzero"`
	Tags        []string `json:"tags" validate:"min=1"`
}

/*
Update a blueprint
*/
func updateBlueprint(u *db.User, r *http.Request) (interface{}, *utils.ErrorResponse) {
	var request PutBlueprintRequest
	e := utils.ValidateRequestBody(r, &request)

	if e != nil {
		return nil, e
	}

	blueprint, e := parseBlueprint(r)

	if e != nil {
		return nil, e
	}

	if blueprint.UserID != u.ID {
		return nil, &utils.Error_no_access
	}

	for _, t := range blueprint.GetTags() {
		bt := db.BlueprintTag{
			BlueprintId: blueprint.ID,
			TagId:       t.ID,
		}

		bt.Delete()
	}

	for _, tag := range request.Tags {
		t := &db.Tag{
			Name: tag,
		}

		t.Save()

		bt := db.BlueprintTag{
			BlueprintId: blueprint.ID,
			TagId:       t.ID,
		}

		bt.Save()
	}

	blueprint.Name = request.Name
	blueprint.Description = request.Description

	blueprint.Save()

	return nil, nil
}

/*
Delete a blueprint
*/
func deleteBlueprint(u *db.User, r *http.Request) (interface{}, *utils.ErrorResponse) {
	blueprint, e := parseBlueprint(r)

	if e != nil {
		return nil, e
	}

	if blueprint.UserID != u.ID {
		return nil, &utils.Error_no_access
	}

	blueprint.Delete()

	return nil, nil
}

type GetRevisionsResponse struct {
	Revisions []*Revision `json:"revisions"`
}

/*
Get all revisions
*/
func getRevisions(r *http.Request) (interface{}, *utils.ErrorResponse) {
	blueprint, e := parseBlueprint(r)

	if e != nil {
		return nil, e
	}

	getComments := len(r.URL.Query()["comments"]) > 0

	revisions := blueprint.GetRevisions()
	reRevision := make([]*Revision, len(revisions))

	authUser := db.GetAuthUser(r)

	for i, revision := range revisions {
		rev, err := revisionToJSON(authUser, &revision, getComments)
		if err != nil {
			return nil, err
		}
		reRevision[i] = rev
	}

	return GetRevisionsResponse{
		Revisions: reRevision,
	}, nil
}

/*
Get latest revision from blueprint
*/
func getRevisionLatest(r *http.Request) (interface{}, *utils.ErrorResponse) {
	blueprint, e := parseBlueprint(r)

	if e != nil {
		return nil, e
	}

	getComments := len(r.URL.Query()["comments"]) > 0

	authUser := db.GetAuthUser(r)
	revision := blueprint.GetLatestRevision()
	if revision == nil || revision.DeletedAt != nil {
		return nil, &utils.Error_revision_not_found
	}

	return revisionToJSON(authUser, revision, getComments)
}

/*
Get specific revision from blueprint
*/
func getRevisionIncremental(r *http.Request) (interface{}, *utils.ErrorResponse) {
	blueprint, e := parseBlueprint(r)

	if e != nil {
		return nil, e
	}

	revisionI, err := strconv.ParseUint(mux.Vars(r)["revision"], 10, 32)
	if err != nil {
		return nil, &utils.Error_revision_not_found
	}

	getComments := len(r.URL.Query()["comments"]) > 0

	authUser := db.GetAuthUser(r)
	revision := blueprint.GetRevision(uint(revisionI))
	return revisionToJSON(authUser, revision, getComments)
}

func revisionToJSON(authUser *db.User, revision *db.Revision, getComments bool) (*Revision, *utils.ErrorResponse) {
	if revision == nil || revision.DeletedAt != nil {
		return nil, &utils.Error_revision_not_found
	}

	ratings := revision.GetRatings()
	thumbsUp, thumbsDown, userVote := 0, 0, 0

	for _, rating := range ratings {
		if rating.ThumbsUp {
			thumbsUp++
		} else {
			thumbsDown++
		}

		if authUser != nil && authUser.ID == rating.UserID {
			if rating.ThumbsUp {
				userVote = 1
			} else {
				userVote = 2
			}
		}
	}

	var reComment []*Comment

	if getComments {
		comments := revision.GetComments()
		reComment = make([]*Comment, len(comments))

		for i, comment := range comments {
			reComment[i] = &Comment{
				Id:        comment.ID,
				UserId:    comment.UserID,
				CreatedAt: comment.CreatedAt,
				UpdatedAt: comment.UpdatedAt,
				Message:   comment.Message,
			}
		}
	}

	blueprintString, err := storage.LoadRevision(revision.ID)

	if err != nil {
		return nil, &utils.Error_internal_error
	}

	return &Revision{
		Id:          revision.ID,
		Revision:    revision.Revision,
		Changes:     revision.Changes,
		CreatedAt:   revision.CreatedAt,
		UpdatedAt:   revision.UpdatedAt,
		BlueprintID: revision.BlueprintID,
		Blueprint:   blueprintString,
		ThumbsUp:    thumbsUp,
		ThumbsDown:  thumbsDown,
		UserVote:    userVote,
		Comments:    reComment,
		Version:     revision.BlueprintVersion,
	}, nil
}

func parseBlueprint(r *http.Request) (*db.Blueprint, *utils.ErrorResponse) {
	blueprintId, err := strconv.ParseUint(mux.Vars(r)["blueprint"], 10, 32)

	if err != nil {
		return nil, &utils.Error_blueprint_not_found
	}

	return findBlueprintById(uint(blueprintId))
}

func findBlueprintById(blueprintId uint) (*db.Blueprint, *utils.ErrorResponse) {
	blueprint := db.GetBlueprintById(uint(blueprintId))

	if blueprint == nil {
		return nil, &utils.Error_blueprint_not_found
	}

	return blueprint, nil
}
