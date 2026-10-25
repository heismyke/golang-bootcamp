package clients

import (
	"context"
	"errors"
	"time"

	"github.com/minio/minio-go/v7"
)

// ═════════════════════════════════════════════════════════════
// PART 5: OBJECT STORAGE (S3)
// Files (images, PDFs, exports, backups) don't go in your database; they
// go in object storage. Locally we run SeaweedFS, which speaks the same API
// as Amazon S3, so this code works against AWS S3, Cloudflare R2, Google
// Cloud Storage etc. too, with different credentials. We use the minio-go
// client, a popular S3 client that works with any S3-compatible server.
// Docs: https://pkg.go.dev/github.com/minio/minio-go/v7
// Credentials for the local server: s3admin / s3secret (see s3.json).
// ═════════════════════════════════════════════════════════════

var ErrNotFound = errors.New("object not found")

// ─── 17. Connecting ───────────────────────────────────────────
// NewS3 creates a client for endpoint ("localhost:59000") with static
// credentials and Secure: false (no TLS locally), then checks it works
// with a cheap call (ListBuckets).
// Hint: minio.New(endpoint, &minio.Options{Creds: credentials.NewStaticV4(key, secret, ""), ...})
func NewS3(ctx context.Context, endpoint, accessKey, secretKey string) (*minio.Client, error) {
	return nil, ErrTODO
}

// ─── 18. Buckets ──────────────────────────────────────────────
// EnsureBucket creates the bucket if it doesn't exist. Safe to call many times.
func EnsureBucket(ctx context.Context, c *minio.Client, bucket string) error {
	return ErrTODO
}

// ─── 19. Upload and download ──────────────────────────────────
// Upload stores data under key with the given content type.
// Hint: PutObject takes an io.Reader (bytes.NewReader) and the size.
func Upload(ctx context.Context, c *minio.Client, bucket, key string, data []byte, contentType string) error {
	return ErrTODO
}

// Download returns the object's bytes, or ErrNotFound if the key doesn't
// exist. Careful: GetObject is lazy and returns no error for a missing key;
// the error only appears when you read (or Stat). Check
// minio.ToErrorResponse(err).Code == "NoSuchKey".
func Download(ctx context.Context, c *minio.Client, bucket, key string) ([]byte, error) {
	return nil, ErrTODO
}

// ─── 20. Presigned URLs and listing ───────────────────────────
// PresignedURL returns a temporary URL anyone can use to download the object
// without credentials, valid for expiry. Apps use this to let browsers
// download private files directly from storage instead of through the API.
func PresignedURL(ctx context.Context, c *minio.Client, bucket, key string, expiry time.Duration) (string, error) {
	return "", ErrTODO
}

// ListKeys returns the keys under prefix, sorted.
// Hint: ListObjects returns a channel; set Recursive: true.
func ListKeys(ctx context.Context, c *minio.Client, bucket, prefix string) ([]string, error) {
	return nil, ErrTODO
}
