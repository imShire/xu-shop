package admin

// UploadSettingsResp 上传设置响应。
type UploadSettingsResp struct {
	Driver           string `json:"driver"`
	PublicBaseURL    string `json:"public_base_url"`
	MaxSizeMB        int    `json:"max_size_mb"`
	LocalDir         string `json:"local_dir"`
	S3Vendor         string `json:"s3_vendor"`
	S3Endpoint       string `json:"s3_endpoint"`
	S3Region         string `json:"s3_region"`
	S3Bucket         string `json:"s3_bucket"`
	S3Prefix         string `json:"s3_prefix"`
	S3ForcePathStyle bool   `json:"s3_force_path_style"`
	S3AccessKeySet   bool   `json:"s3_access_key_set"`
	S3SecretKeySet   bool   `json:"s3_secret_key_set"`
}

// UpdateUploadSettingsReq 更新上传设置请求。
type UpdateUploadSettingsReq struct {
	Driver           string `json:"driver" binding:"required,oneof=local s3"`
	PublicBaseURL    string `json:"public_base_url" binding:"required,url"`
	MaxSizeMB        int    `json:"max_size_mb" binding:"required,gte=1,lte=50"`
	LocalDir         string `json:"local_dir"`
	S3Vendor         string `json:"s3_vendor"`
	S3Endpoint       string `json:"s3_endpoint"`
	S3Region         string `json:"s3_region"`
	S3Bucket         string `json:"s3_bucket"`
	S3AccessKey      string `json:"s3_access_key"`
	S3SecretKey      string `json:"s3_secret_key"`
	S3Prefix         string `json:"s3_prefix"`
	S3ForcePathStyle bool   `json:"s3_force_path_style"`
}
