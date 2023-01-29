package aliyundrive

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"io"
	"math/big"
	"net/http"
	"time"
)

const RootFileId = "root"
const KB = 1024
const MB = 1024 * KB
const GB = 1024 * MB
const TB = 1024 * GB
const LimitMax = 200

const OrderDirectionDesc = "DESC"
const OrderDirectionAsc = "ASC"
const OrderByName = "name"
const OrderByUpdatedAt = "updated_at"
const OrderByCreatedAt = "created_at"

func GetProofStart(accessToken string, size uint64) uint64 {
	hash := md5.Sum([]byte(accessToken))
	bigInt := new(big.Int).SetBytes(hash[:8])
	return new(big.Int).Mod(bigInt, new(big.Int).SetUint64(size)).Uint64()
}

type Item struct {
	FileId          string    `json:"file_id"`
	Name            string    `json:"name"`
	ParentFileId    string    `json:"parent_file_id"`
	Type            string    `json:"type"`
	Starred         bool      `json:"starred"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	TrashedAt       time.Time `json:"trashed_at"`
	GMTExpired      time.Time `json:"gmt_expired"`
	Category        string    `json:"category"`
	ContentHash     string    `json:"content_hash"`
	Size            uint64    `json:"size"`
	UserMeta        string    `json:"user_meta"`
	FileExtension   string    `json:"file_extension"`
	MimeType        string    `json:"mime_type"`
	PunishFlag      int       `json:"punish_flag"`
	Thumbnail       string    `json:"thumbnail"`
	Url             string    `json:"url"`
	Hidden          bool      `json:"hidden"`
	Trashed         bool      `json:"trashed"`
	Status          string    `json:"status"`
	EncryptMode     string    `json:"encrypt_mode"`
	ContentHashName string    `json:"content_hash_name"`
	ContentType     string    `json:"content_type"`
	Crc64Hash       string    `json:"crc64_hash"`
	MimeExtension   string    `json:"mime_extension"`
	DownloadUrl     string    `json:"download_url"`
	UploadId        string    `json:"upload_id"`
	Labels          []string  `json:"labels"`
}

type ItemQuery []*Item

func (q ItemQuery) ByName(name string) (item *Item, exists bool) {
	for _, i := range q {
		if i.Name == name {
			item = i
			exists = true
			break
		}
	}
	return
}

type GetPersonalInfoRequest struct {
}

type GetPersonalInfoResponse struct {
	PersonalRightsInfo *struct {
		Name       string `json:"name"`
		SpuId      string `json:"spu_id"`
		IsExpires  bool   `json:"is_expires"`
		Privileges []*struct {
			FeatureId     string `json:"feature_id"`
			FeatureAttrId string `json:"feature_attr_id"`
			Quota         int    `json:"quota"`
		} `json:"privileges"`
	} `json:"personal_rights_info"`
	PersonalSpaceInfo *struct {
		TotalSize uint64 `json:"total_size"`
		UsedSize  uint64 `json:"used_size"`
	} `json:"personal_space_info"`
}

func (c *Drive) DoGetPersonalInfoRequest(ctx context.Context, request GetPersonalInfoRequest) (*GetPersonalInfoResponse, error) {
	resp, err := c.requestWithCredit(ctx, "https://api.aliyundrive.com/v2/databox/get_personal_info", Object{})
	if err != nil {
		return nil, err
	}

	result := new(GetPersonalInfoResponse)
	err = json.Unmarshal(resp, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

type ListRequest struct {
	ParentFileId   string `json:"parent_file_id,omitempty"`
	OrderBy        string `json:"order_by,omitempty"`
	OrderDirection string `json:"order_direction,omitempty"`
	Limit          int    `json:"limit,omitempty"`
	NextMarker     string `json:"marker,omitempty"`
}
type ListResponse struct {
	Items      []*Item `json:"items"`
	NextMarker string  `json:"next_marker"`
}

func (c *Drive) DoListRequest(ctx context.Context, request ListRequest) (*ListResponse, error) {
	params := &struct {
		DriveId string `json:"drive_id"`
		Fields  string `json:"fields"`
		ListRequest
	}{
		DriveId:     c.driveId,
		Fields:      "*",
		ListRequest: request,
	}
	resp, err := c.requestWithCredit(ctx, "https://api.aliyundrive.com/adrive/v3/file/list", params)
	if err != nil {
		return nil, err
	}

	result := new(ListResponse)
	err = json.Unmarshal(resp, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

type SearchRequest struct {
	Name           string `json:"-"`
	OrderBy        string `json:"order_by,omitempty"`
	OrderDirection string `json:"order_direction,omitempty"`
	Limit          int    `json:"limit,omitempty"`
	NextMarker     string `json:"marker,omitempty"`
}

type SearchResponse struct {
	Items      []*Item `json:"items"`
	NextMarker string  `json:"next_marker"`
}

func (c *Drive) DoSearchRequest(ctx context.Context, request SearchRequest) (*SearchResponse, error) {

	params := &struct {
		DriveId string `json:"drive_id"`
		Query   string `json:"query"`
		SearchRequest
	}{
		DriveId:       c.driveId,
		SearchRequest: request,
	}
	params.Query = `name match "` + params.Name + `"`

	resp, err := c.requestWithCredit(ctx, "https://api.aliyundrive.com/adrive/v3/file/search", params)
	if err != nil {
		return nil, err
	}

	result := new(SearchResponse)
	err = json.Unmarshal(resp, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

type GetRequest struct {
	FileId string `json:"file_id"`
}

type GetResponse struct {
	Item
}

func (c *Drive) DoGetRequest(ctx context.Context, request GetRequest) (*GetResponse, error) {

	params := &struct {
		DriveId string `json:"drive_id"`
		GetRequest
	}{
		DriveId:    c.driveId,
		GetRequest: request,
	}
	resp, err := c.requestWithCredit(ctx, "https://api.aliyundrive.com/v2/file/get", params)
	if err != nil {
		return nil, err
	}

	result := new(GetResponse)
	err = json.Unmarshal(resp, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

type GetDownloadUrlRequest struct {
	FileId string `json:"file_id"`
}

type GetDownloadUrlResponse struct {
	FileId          string    `json:"file_id"`
	Size            uint64    `json:"size"`
	ContentHash     string    `json:"content_hash"`
	ContentHashName string    `json:"content_hash_name"`
	Crc64Hash       string    `json:"crc64_hash"`
	Expiration      time.Time `json:"expiration"`
	InternalUrl     string    `json:"internal_url"`
	Url             string    `json:"url"`
}

func (c *Drive) DoGetDownloadUrlRequest(ctx context.Context, request GetDownloadUrlRequest) (*GetDownloadUrlResponse, error) {

	params := &struct {
		DriveId string `json:"drive_id"`
		GetDownloadUrlRequest
	}{
		DriveId:               c.driveId,
		GetDownloadUrlRequest: request,
	}
	resp, err := c.requestWithCredit(ctx, "https://api.aliyundrive.com/v2/file/get_download_url", params)
	if err != nil {
		return nil, err
	}

	result := new(GetDownloadUrlResponse)
	err = json.Unmarshal(resp, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

type GetFolderSizeInfoRequest struct {
	FileId string `json:"file_id"`
}

type GetFolderSizeInfoResponse struct {
	FileCount   uint64 `json:"file_count"`
	FolderCount uint64 `json:"folder_count"`
	Size        uint64 `json:"size"`
}

func (c *Drive) DoGetFolderSizeInfoRequest(ctx context.Context, request GetFolderSizeInfoRequest) (*GetFolderSizeInfoResponse, error) {

	params := &struct {
		DriveId string `json:"drive_id"`
		GetFolderSizeInfoRequest
	}{
		DriveId:                  c.driveId,
		GetFolderSizeInfoRequest: request,
	}
	resp, err := c.requestWithCredit(ctx, "https://api.aliyundrive.com/adrive/v1/file/get_folder_size_info", params)
	if err != nil {
		return nil, err
	}

	result := new(GetFolderSizeInfoResponse)
	err = json.Unmarshal(resp, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

type CreateFolderRequest struct {
	Name         string `json:"name"`
	ParentFileId string `json:"parent_file_id"`
}

type CreateFolderResponse struct {
	FileId       string `json:"file_id"`
	FileName     string `json:"file_name"`
	ParentFileId string `json:"parent_file_id"`
	Type         string `json:"type"`
	EncryptMode  string `json:"encrypt_mode"`
}

func (c *Drive) DoCreateFolderRequest(ctx context.Context, request CreateFolderRequest) (*CreateFolderResponse, error) {
	params := &struct {
		DriveId       string `json:"drive_id"`
		CheckNameMode string `json:"check_name_mode"`
		Type          string `json:"type"`
		CreateFolderRequest
	}{
		DriveId:             c.driveId,
		CheckNameMode:       "refuse",
		Type:                "folder",
		CreateFolderRequest: request,
	}

	resp, err := c.requestWithCredit(ctx, "https://api.aliyundrive.com/adrive/v2/file/createWithFolders", params)
	if err != nil {
		return nil, err
	}

	result := new(CreateFolderResponse)
	err = json.Unmarshal(resp, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

type CreateFileRequest struct {
	Name         string `json:"name"`
	ParentFileId string `json:"parent_file_id"`
	Size         uint64 `json:"size"`
	PreHash      string `json:"pre_hash"`
	ChunkSize    uint64 `json:"-"`
}

type CreateFileResponse struct {
	FileId       string `json:"file_id"`
	FileName     string `json:"file_name"`
	ParentFileId string `json:"parent_file_id"`
	RapidUpload  bool   `json:"rapid_upload"`
	Type         string `json:"type"`
	EncryptMode  string `json:"encrypt_mode"`
	UploadId     string `json:"upload_id"`
	PartInfoList []*struct {
		PartNumber        int    `json:"part_number"`
		ContentType       string `json:"content_type"`
		InternalUploadUrl string `json:"internal_upload_url"`
		UploadUrl         string `json:"upload_url"`
	} `json:"part_info_list"`
}

func (c *Drive) DoCreateFileRequest(ctx context.Context, request CreateFileRequest) (*CreateFileResponse, error) {
	params := &struct {
		DriveId       string `json:"drive_id"`
		DeviceName    string `json:"device_name"`
		CreateScene   string `json:"create_scene"`
		CheckNameMode string `json:"check_name_mode"`
		Type          string `json:"type"`
		PartInfoList  Array  `json:"part_info_list"`
		CreateFileRequest
	}{
		DriveId:           c.driveId,
		CheckNameMode:     "auto_rename",
		CreateScene:       "file_upload",
		Type:              "file",
		CreateFileRequest: request,
	}

	var partCount int
	if params.ChunkSize == 0 {
		partCount = 1
	} else {
		partCount = int(params.Size / params.ChunkSize)
		if params.Size%params.ChunkSize > 0 {
			partCount++
		}
	}

	params.PartInfoList = make(Array, partCount)
	for i := 0; i < partCount; i++ {
		params.PartInfoList[i] = Object{"part_number": i}
	}

	resp, err := c.requestWithCredit(ctx, "https://api.aliyundrive.com/adrive/v2/file/createWithFolders", params)
	if err != nil {
		return nil, err
	}

	result := new(CreateFileResponse)
	err = json.Unmarshal(resp, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

type DownloadFileRequest struct {
	Url    string
	Header http.Header
}

type DownloadFileResponse struct {
	Reader io.ReadCloser
}

func (c *Drive) DoDownloadFileRequest(ctx context.Context, request DownloadFileRequest) (*DownloadFileResponse, error) {
	httpRequest, err := http.NewRequestWithContext(ctx, "GET", request.Url, nil)
	if err != nil {
		return nil, err
	}
	for k := range request.Header {
		httpRequest.Header.Add(k, request.Header.Get(k))
	}
	httpRequest.Header.Set("Origin", "https://www.aliyundrive.com")
	httpRequest.Header.Set("Referer", "https://www.aliyundrive.com/")

	resp, err := c.httpClient.Do(httpRequest)
	if err != nil {
		return nil, err
	}
	return &DownloadFileResponse{
		Reader: resp.Body,
	}, nil
}

type UploadFileRequest struct {
	Url  string
	File io.Reader
}

type UploadFileResponse struct {
}

func (c *Drive) DoUploadFileRequest(ctx context.Context, request UploadFileRequest) (*UploadFileResponse, error) {
	httpRequest, err := http.NewRequestWithContext(ctx, "PUT", request.Url, request.File)
	if err != nil {
		return nil, err
	}
	httpRequest.Header.Set("Origin", "https://www.aliyundrive.com")
	httpRequest.Header.Set("Referer", "https://www.aliyundrive.com/")

	resp, err := c.httpClient.Do(httpRequest)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}
	return &UploadFileResponse{}, nil
}

type CompleteUploadFileRequest struct {
	FileId   string `json:"file_id"`
	UploadId string `json:"upload_id"`
}

type CompleteUploadFileResponse struct {
	Item
}

func (c *Drive) DoCompleteUploadFileRequest(ctx context.Context, request CompleteUploadFileRequest) (*CompleteUploadFileResponse, error) {
	params := &struct {
		DriveId string `json:"drive_id"`
		CompleteUploadFileRequest
	}{
		DriveId:                   c.driveId,
		CompleteUploadFileRequest: request,
	}

	resp, err := c.requestWithCredit(ctx, "https://api.aliyundrive.com/v2/file/complete", params)
	if err != nil {
		return nil, err
	}

	result := new(CompleteUploadFileResponse)
	err = json.Unmarshal(resp, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

type RapidCreateFileRequest struct {
	Name         string `json:"name"`
	ParentFileId string `json:"parent_file_id"`
	Size         uint64 `json:"size"`
	ChunkSize    uint64 `json:"-"`
	ContentHash  string `json:"content_hash"`
	ProofCode    string `json:"proof_code"`
	AccessToken  string `json:"-"`
}

type RapidCreateFileResponse struct {
	FileId       string `json:"file_id"`
	FileName     string `json:"file_name"`
	ParentFileId string `json:"parent_file_id"`
	RapidUpload  bool   `json:"rapid_upload"`
	Type         string `json:"type"`
	EncryptMode  string `json:"encrypt_mode"`
	UploadId     string `json:"upload_id"`
}

func (c *Drive) DoRapidCreateFileRequest(ctx context.Context, request RapidCreateFileRequest) (*RapidCreateFileResponse, error) {
	params := &struct {
		DriveId         string `json:"drive_id"`
		DeviceName      string `json:"device_name"`
		CreateScene     string `json:"create_scene"`
		CheckNameMode   string `json:"check_name_mode"`
		ContentHashName string `json:"content_hash_name"`
		Type            string `json:"type"`
		ProofVersion    string `json:"proof_version"`
		PartInfoList    Array  `json:"part_info_list"`
		RapidCreateFileRequest
	}{
		DriveId:                c.driveId,
		CheckNameMode:          "auto_rename",
		CreateScene:            "file_upload",
		ContentHashName:        "sha1",
		Type:                   "file",
		ProofVersion:           "v1",
		RapidCreateFileRequest: request,
	}

	var partCount int
	if params.ChunkSize == 0 {
		partCount = 1
	} else {
		partCount = int(params.Size / params.ChunkSize)
		if params.Size%params.ChunkSize > 0 {
			partCount++
		}
	}

	params.PartInfoList = make(Array, partCount)
	for i := 0; i < partCount; i++ {
		params.PartInfoList[i] = Object{"part_number": i}
	}

	httpRequest, err := c.toRequest(ctx, "https://api.aliyundrive.com/adrive/v2/file/createWithFolders", params)
	if err != nil {
		return nil, err
	}
	httpRequest.Header.Set("Authorization", "Bearer "+params.AccessToken)
	resp, err := c.doRequest(httpRequest)
	if err != nil {
		return nil, err
	}

	result := new(RapidCreateFileResponse)
	err = json.Unmarshal(resp, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

type RenameRequest struct {
	FileId string `json:"file_id"`
	Name   string `json:"name"`
}

type RenameResponse struct {
	Item
}

func (c *Drive) DoRenameRequest(ctx context.Context, request RenameRequest) (*RenameResponse, error) {
	params := &struct {
		DriveId       string `json:"drive_id"`
		CheckNameMode string `json:"check_name_mode"`
		RenameRequest
	}{
		DriveId:       c.driveId,
		CheckNameMode: "refuse",
		RenameRequest: request,
	}

	resp, err := c.requestWithCredit(ctx, "https://api.aliyundrive.com/v3/file/update", params)
	if err != nil {
		return nil, err
	}

	result := new(RenameResponse)
	err = json.Unmarshal(resp, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

type MoveRequest struct {
	FileId         string `json:"file_id"`
	ToParentFileId string `json:"to_parent_file_id"`
}

type MoveResponse struct {
	FileId string `json:"file_id"`
}

func (c *Drive) DoMoveRequest(ctx context.Context, request MoveRequest) (*MoveResponse, error) {
	params := &struct {
		DriveId   string `json:"drive_id"`
		ToDriveId string `json:"to_drive_id"`
		MoveRequest
	}{
		DriveId:     c.driveId,
		ToDriveId:   c.driveId,
		MoveRequest: request,
	}

	resp, err := c.requestWithCredit(ctx, "https://api.aliyundrive.com/v3/file/move", params)
	if err != nil {
		return nil, err
	}

	result := new(MoveResponse)
	err = json.Unmarshal(resp, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

type TrashRequest struct {
	FileId string `json:"file_id"`
}

type TrashResponse struct {
	AsyncTaskId string `json:"async_task_id"`
	FileId      string `json:"file_id"`
}

func (c *Drive) DoTrashRequest(ctx context.Context, request TrashRequest) (*TrashResponse, error) {
	params := &struct {
		DriveId string `json:"drive_id"`
		TrashRequest
	}{
		DriveId:      c.driveId,
		TrashRequest: request,
	}

	resp, err := c.requestWithCredit(ctx, "https://api.aliyundrive.com/v2/recyclebin/trash", params)
	if err != nil {
		return nil, err
	}

	result := new(TrashResponse)
	err = json.Unmarshal(resp, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

type ClearTrashRequest struct{}

type ClearTrashResponse struct {
	AsyncTaskId string `json:"async_task_id"`
	TaskId      string `json:"task_id"`
}

func (c *Drive) DoClearTrashRequest(ctx context.Context, request ClearTrashRequest) (*ClearTrashResponse, error) {
	params := &struct {
		DriveId string `json:"drive_id"`
		ClearTrashRequest
	}{
		DriveId:           c.driveId,
		ClearTrashRequest: request,
	}

	resp, err := c.requestWithCredit(ctx, "https://api.aliyundrive.com/v2/recyclebin/clear", params)
	if err != nil {
		return nil, err
	}

	result := new(ClearTrashResponse)
	err = json.Unmarshal(resp, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

type ListTrashRequest struct {
	OrderBy        string `json:"order_by,omitempty"`
	OrderDirection string `json:"order_direction,omitempty"`
	Limit          int    `json:"limit,omitempty"`
	NextMarker     string `json:"marker,omitempty"`
}

type ListTrashResponse struct {
	Items      []*Item `json:"items"`
	NextMarker string  `json:"next_marker"`
}

func (c *Drive) DoListTrashRequest(ctx context.Context, request ListTrashRequest) (*ListTrashResponse, error) {
	params := &struct {
		DriveId string `json:"drive_id"`
		ListTrashRequest
	}{
		DriveId:          c.driveId,
		ListTrashRequest: request,
	}

	resp, err := c.requestWithCredit(ctx, "https://api.aliyundrive.com/adrive/v2/recyclebin/list", params)
	if err != nil {
		return nil, err
	}

	result := new(ListTrashResponse)
	err = json.Unmarshal(resp, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

type RestoreRequest struct {
	FileId string `json:"file_id"`
}

type RestoreResponse struct {
}

func (c *Drive) DoRestoreRequest(ctx context.Context, request RestoreRequest) (*RestoreResponse, error) {
	params := &struct {
		DriveId string `json:"drive_id"`
		RestoreRequest
	}{
		DriveId:        c.driveId,
		RestoreRequest: request,
	}

	accessToken, err := c.tokenManager.AccessToken(ctx)
	if err != nil {
		return nil, err
	}
	httpRequest, err := c.toRequest(ctx, "https://api.aliyundrive.com/v2/recyclebin/restore", params)
	if err != nil {
		return nil, err
	}
	httpRequest.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := c.httpClient.Do(httpRequest)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}
	return &RestoreResponse{}, nil
}

type DeleteRequest struct {
	FileId string `json:"file_id"`
}

type DeleteResponse struct {
}

func (c *Drive) DoDeleteRequest(ctx context.Context, request DeleteRequest) (*DeleteResponse, error) {
	params := &struct {
		DriveId string `json:"drive_id"`
		DeleteRequest
	}{
		DriveId:       c.driveId,
		DeleteRequest: request,
	}

	accessToken, err := c.tokenManager.AccessToken(ctx)
	if err != nil {
		return nil, err
	}
	httpRequest, err := c.toRequest(ctx, "https://api.aliyundrive.com/v3/file/delete", params)
	if err != nil {
		return nil, err
	}
	httpRequest.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := c.httpClient.Do(httpRequest)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}
	return &DeleteResponse{}, nil
}
