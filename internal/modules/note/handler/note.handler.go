package handler

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/sammidev/goca/internal/modules/note/dto"
	"github.com/sammidev/goca/internal/pkg/response"
	"github.com/sammidev/goca/internal/server/api/middleware"
)

type NoteHandler struct {
	noteService NoteService
}

func NewNoteHandler(noteService NoteService) *NoteHandler {
	return &NoteHandler{
		noteService: noteService,
	}
}

// CreateNote godoc
//
//	@Summary		Create a new note
//	@Description	Create a new note for the authenticated user
//	@Tags			notes
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.CreateNoteRequest							true	"Note creation data"
//	@Success		201		{object}	response.Response{data=dto.CreateNoteResponse}	"Note created successfully"
//	@Failure		400		{object}	response.Response
//	@Failure		401		{object}	response.Response
//	@Failure		500		{object}	response.Response
//	@Router			/notes [post]
func (h *NoteHandler) CreateNote(c *fiber.Ctx) error {
	var req dto.CreateNoteRequest
	if err := c.BodyParser(&req); err != nil {
		return response.HandleErrorAPI(c, err)
	}

	req.UserID = middleware.GetUser(c).UserID

	res, err := h.noteService.CreateNote(c.UserContext(), &req)
	if err != nil {
		return response.HandleErrorAPI(c, err)
	}

	return response.HandleSuccessAPI(c, http.StatusCreated, "Note created successfully", res, nil)
}

// GetNote godoc
//
//	@Summary		Get a note by ID
//	@Description	Retrieve a specific note by its ID for the authenticated user
//	@Tags			notes
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string										true	"Note ID"
//	@Success		200	{object}	response.Response{data=dto.GetNoteResponse}	"Note retrieved successfully"
//	@Failure		400	{object}	response.Response
//	@Failure		401	{object}	response.Response
//	@Failure		404	{object}	response.Response
//	@Failure		500	{object}	response.Response
//	@Router			/notes/{id} [get]
func (h *NoteHandler) GetNote(c *fiber.Ctx) error {
	noteID, err := ParseUUIDParam(c, "id")
	if err != nil {
		return response.HandleErrorAPI(c, err)
	}

	req := dto.GetNoteRequest{
		NoteID: noteID,
		UserID: middleware.GetUser(c).UserID,
	}

	res, err := h.noteService.GetNote(c.UserContext(), &req)
	if err != nil {
		return response.HandleErrorAPI(c, err)
	}

	return response.HandleSuccessAPI(c, http.StatusOK, "Note retrieved successfully", res, nil)
}

// UpdateNote godoc
//
//	@Summary		Update a note
//	@Description	Update an existing note for the authenticated user
//	@Tags			notes
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string											true	"Note ID"
//	@Param			request	body		dto.UpdateNoteRequest							true	"Note update data"
//	@Success		200		{object}	response.Response{data=dto.UpdateNoteResponse}	"Note updated successfully"
//	@Failure		400		{object}	response.Response
//	@Failure		401		{object}	response.Response
//	@Failure		404		{object}	response.Response
//	@Failure		500		{object}	response.Response
//	@Router			/notes/{id} [put]
func (h *NoteHandler) UpdateNote(c *fiber.Ctx) error {
	noteID, err := ParseUUIDParam(c, "id")
	if err != nil {
		return response.HandleErrorAPI(c, err)
	}

	var req dto.UpdateNoteRequest
	if err := c.BodyParser(&req); err != nil {
		return response.HandleErrorAPI(c, err)
	}

	req.UserID = middleware.GetUser(c).UserID
	req.NoteID = noteID

	res, err := h.noteService.UpdateNote(c.UserContext(), &req)
	if err != nil {
		return response.HandleErrorAPI(c, err)
	}

	return response.HandleSuccessAPI(c, http.StatusOK, "Note updated successfully", res, nil)
}

// DeleteNote godoc
//
//	@Summary		Delete a note
//	@Description	Delete a note by its ID for the authenticated user
//	@Tags			notes
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string				true	"Note ID"
//	@Success		200	{object}	response.Response	"Note deleted successfully"
//	@Failure		400	{object}	response.Response
//	@Failure		401	{object}	response.Response
//	@Failure		404	{object}	response.Response
//	@Failure		500	{object}	response.Response
//	@Router			/notes/{id} [delete]
func (h *NoteHandler) DeleteNote(c *fiber.Ctx) error {
	noteID, err := ParseUUIDParam(c, "id")
	if err != nil {
		return response.HandleErrorAPI(c, err)
	}

	req := dto.DeleteNoteRequest{
		NoteID: noteID,
		UserID: middleware.GetUser(c).UserID,
	}

	err = h.noteService.DeleteNote(c.UserContext(), &req)
	if err != nil {
		return response.HandleErrorAPI(c, err)
	}

	return response.HandleSuccessAPI(c, http.StatusOK, "Note deleted successfully", nil, nil)
}

// GetNotes godoc
//
//	@Summary		List notes
//	@Description	List all notes for the authenticated user with pagination and filtering
//	@Tags			notes
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page			query		int												false	"Page number"		default(1)
//	@Param			per_page		query		int												false	"Items per page"	default(10)
//	@Param			keyword			query		string											false	"Search keyword"
//	@Param			sort_by			query		string											false	"Sort by column (e.g., created_at)"
//	@Param			sort_direction	query		string											false	"Sort direction (asc/desc)"	default(asc)
//	@Success		200				{object}	response.Response{data=dto.GetNotesResponse}	"Notes listed successfully"
//	@Failure		400				{object}	response.Response
//	@Failure		401				{object}	response.Response
//	@Failure		500				{object}	response.Response
//	@Router			/notes [get]
func (h *NoteHandler) GetNotes(c *fiber.Ctx) error {
	req := dto.NewGetNotesRequest()
	if err := c.QueryParser(req); err != nil {
		return response.HandleErrorAPI(c, err)
	}

	req.UserID = middleware.GetUser(c).UserID

	res, err := h.noteService.GetNotes(c.UserContext(), req)
	if err != nil {
		return response.HandleErrorAPI(c, err)
	}

	return response.HandleSuccessAPI(c, http.StatusOK, "Notes listed successfully", res.List, res.Paging)
}
