package schema

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testBodyStruct is a simple struct for JSON body tests.
type testBodyStruct struct {
	Name  string `schema:"name"`
	Count int    `schema:"count"`
}

func createBodyMetadata(bodyType BodyType, fieldType reflect.Type) *StructMetadata {
	var tagValue string
	switch bodyType {
	case BodyTypeMultipart:
		tagValue = `body:"multipart"`
	case BodyTypeFile:
		tagValue = `body:"file"`
	case BodyTypeStructured:
		tagValue = `body:""`
	default:
		tagValue = `body:""`
	}
	structType := reflect.StructOf([]reflect.StructField{
		{
			Name: goFieldBody,
			Type: fieldType,
			Tag:  reflect.StructTag(tagValue),
		},
	})
	metadata := NewDefaultMetadata()
	structMeta, err := metadata.GetStructMetadata(structType)
	if err != nil {
		panic(fmt.Sprintf("failed to get struct metadata: %v", err))
	}

	return structMeta
}

func createBodyRequest(contentType string, body []byte) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(body))
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	return req
}

func TestDecoder_DecodeBody_JSON(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		contentType string
		want        map[string]any
		wantErr     bool
	}{
		{
			name:        "valid JSON object",
			body:        `{"name": "test", "count": 42}`,
			contentType: mimeJSON,
			want: map[string]any{
				goFieldBody: map[string]any{
					keyName: valueTest,
					"count": float64(42), // JSON numbers are float64
				},
			},
			wantErr: false,
		},
		{
			name:        "nested objects",
			body:        `{"user": {"name": "john", "age": 30}, "address": {"city": "NYC", "country": "USA"}}`,
			contentType: mimeJSON,
			want: map[string]any{
				goFieldBody: map[string]any{
					keyUser: map[string]any{
						keyName: valueJohnLower,
						keyAge:  float64(30),
					},
					"address": map[string]any{
						keyCity:   valueNYC,
						"country": "USA",
					},
				},
			},
			wantErr: false,
		},
		{
			name:        "deeply nested objects",
			body:        `{"level1": {"level2": {"level3": {"value": "deep"}}}}`,
			contentType: mimeJSON,
			want: map[string]any{
				goFieldBody: map[string]any{
					"level1": map[string]any{
						"level2": map[string]any{
							"level3": map[string]any{
								valuePlain: "deep",
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name:        "JSON with array field",
			body:        `{"items": [1, 2, 3], "tags": ["a", "b"]}`,
			contentType: mimeJSON,
			want: map[string]any{
				goFieldBody: map[string]any{
					keyItems: []any{float64(1), float64(2), float64(3)},
					keyTags:  []any{"a", "b"},
				},
			},
			wantErr: false,
		},
		{
			name:        "JSON array as body",
			body:        `[1, 2, 3]`,
			contentType: mimeJSON,
			want: map[string]any{
				goFieldBody: []any{float64(1), float64(2), float64(3)},
			},
			wantErr: false,
		},
		{
			name:        "JSON string as body",
			body:        `"hello"`,
			contentType: mimeJSON,
			want: map[string]any{
				goFieldBody: "hello",
			},
			wantErr: false,
		},
		{
			name:        "JSON number as body",
			body:        `42`,
			contentType: mimeJSON,
			want: map[string]any{
				goFieldBody: float64(42),
			},
			wantErr: false,
		},
		{
			name:        "empty body returns empty map",
			body:        "",
			contentType: mimeJSON,
			want:        map[string]any{},
			wantErr:     false,
		},
		{
			name:        "invalid JSON",
			body:        `{"name": "test"`,
			contentType: mimeJSON,
			want:        nil,
			wantErr:     true,
		},
	}

	decoder := newTestDecoder()
	metadata := createBodyMetadata(BodyTypeStructured, reflect.TypeFor[testBodyStruct]())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createBodyRequest(tt.contentType, []byte(tt.body))

			result, err := decoder.Decode(req, nil, metadata)

			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestDecoder_DecodeBody_File(t *testing.T) {
	tests := []struct {
		name        string
		body        []byte
		contentType string
		wantData    []byte
	}{
		{
			name:        "binary file content",
			body:        []byte{0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE, 0xFD},
			contentType: mimeOctetStream,
			wantData:    []byte{0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE, 0xFD},
		},
		{
			name:        "text file as binary",
			body:        []byte("Hello, World!"),
			contentType: mimeOctetStream,
			wantData:    []byte("Hello, World!"),
		},
		{
			name:        "empty file",
			body:        []byte{},
			contentType: mimeOctetStream,
			wantData:    nil, // empty body returns empty map
		},
		{
			name:        "large binary content",
			body:        bytes.Repeat([]byte{0xAB, 0xCD}, 1024),
			contentType: mimeOctetStream,
			wantData:    bytes.Repeat([]byte{0xAB, 0xCD}, 1024),
		},
	}

	decoder := newTestDecoder()
	metadata := createBodyMetadata(BodyTypeFile, reflect.TypeFor[[]byte]())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := createBodyRequest(tt.contentType, tt.body)

			result, err := decoder.Decode(req, nil, metadata)

			require.NoError(t, err)

			if tt.wantData == nil {
				// Empty body case
				assert.Empty(t, result)

				return
			}

			// File content is stored under the MapKey "Body"
			data, ok := result[goFieldBody]
			require.True(t, ok, "expected Body key in result")

			//nolint:forcetypeassert // Test code - safe to assert
			assert.Equal(t, tt.wantData, data.([]byte))
		})
	}
}

func TestDecoder_DecodeBody_FileWithReader(t *testing.T) {
	decoder := newTestDecoder()
	metadata := createBodyMetadata(BodyTypeFile, reflect.TypeFor[io.ReadCloser]())

	body := []byte("file content for reader test")
	req := createBodyRequest(mimeOctetStream, body)

	result, err := decoder.Decode(req, nil, metadata)

	require.NoError(t, err)

	data, ok := result[goFieldBody]
	require.True(t, ok, "expected Body key in result")

	//nolint:forcetypeassert // Test code - safe to assert
	assert.Equal(t, body, data.([]byte))
}

func TestDecoder_DecodeBody_NoBodyField(t *testing.T) {
	decoder := newTestDecoder()

	// Metadata without body field
	structType := reflect.StructOf([]reflect.StructField{
		{
			Name: goFieldName,
			Type: reflect.TypeFor[string](),
			Tag:  reflect.StructTag(`schema:"name,location=query"`),
		},
	})
	metadata := NewDefaultMetadata()
	structMeta, err := metadata.GetStructMetadata(structType)
	require.NoError(t, err)

	body := []byte(`{"ignored": "data"}`)
	req := createBodyRequest(mimeJSON, body)

	result, err := decoder.Decode(req, nil, structMeta)

	require.NoError(t, err)
	// Body should be ignored, result should be empty (no query params)
	assert.Empty(t, result)
}

// multipartFormStruct is used for multipart form tests.
type multipartFormStruct struct {
	Name    string // no tag - uses field name "Name"
	Email   string `schema:"email"`
	Age     string `schema:",required"` // no name, uses field name "Age"
	Ignored string `schema:"-"`
}

// multipartFileStruct is used for multipart file upload tests.
type multipartFileStruct struct {
	Name string        `schema:"name"`
	File io.ReadCloser `schema:"file"`
}

// multipartMultiFileStruct is used for multiple file upload tests.
type multipartMultiFileStruct struct {
	Files []io.ReadCloser `schema:"files"`
}

// multipartMixedStruct is used for mixed content tests.
type multipartMixedStruct struct {
	Title string        `schema:"title"`
	File  io.ReadCloser `schema:"attachment"`
}

type multipartFormData struct {
	fields map[string][]string
	files  map[string][]fileData
}

type fileData struct {
	filename string
	content  []byte
}

func createMultipartRequest(data multipartFormData) *http.Request {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add form fields
	for name, values := range data.fields {
		for _, value := range values {
			_ = writer.WriteField(name, value)
		}
	}

	// Add files
	for fieldName, files := range data.files {
		for _, f := range files {
			part, _ := writer.CreateFormFile(fieldName, f.filename)
			_, _ = part.Write(f.content)
		}
	}

	_ = writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/test", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req
}

func TestDecoder_DecodeBody_Multipart(t *testing.T) {
	tests := []struct {
		name      string
		formData  multipartFormData
		bodyType  reflect.Type
		want      map[string]any
		wantErr   bool
		checkFunc func(t *testing.T, result map[string]any) // for file content checks
	}{
		{
			name: "single text field",
			formData: multipartFormData{
				fields: map[string][]string{
					goFieldName: {valueJohn},
				},
			},
			bodyType: reflect.TypeFor[multipartFormStruct](),
			want: map[string]any{
				goFieldBody: map[string]any{
					goFieldName: valueJohn,
				},
			},
		},
		{
			name: "multiple text fields",
			formData: multipartFormData{
				fields: map[string][]string{
					goFieldName: {valueJohn},
					keyEmail:    {valueEmailJohn},
					goFieldAge:  {"30"},
				},
			},
			bodyType: reflect.TypeFor[multipartFormStruct](),
			want: map[string]any{
				goFieldBody: map[string]any{
					goFieldName: valueJohn,      // No tag, uses field name
					keyEmail:    valueEmailJohn, // Tag name "email", not field name "Email"
					goFieldAge:  "30",           // No name in tag, uses field name
				},
			},
		},
		{
			name: "field with multiple values",
			formData: multipartFormData{
				fields: map[string][]string{
					goFieldName: {valueJohn, valueJane},
				},
			},
			bodyType: reflect.TypeFor[multipartFormStruct](),
			want: map[string]any{
				goFieldBody: map[string]any{
					goFieldName: []any{valueJohn, valueJane},
				},
			},
		},
		{
			name: "empty form",
			formData: multipartFormData{
				fields: map[string][]string{},
			},
			bodyType: reflect.TypeFor[multipartFormStruct](),
			want: map[string]any{
				goFieldBody: map[string]any{},
			},
		},
		{
			name: "unknown fields ignored",
			formData: multipartFormData{
				fields: map[string][]string{
					goFieldName:  {valueJohn},
					valueUnknown: {"should be ignored"},
				},
			},
			bodyType: reflect.TypeFor[multipartFormStruct](),
			want: map[string]any{
				goFieldBody: map[string]any{
					goFieldName: valueJohn,
				},
			},
		},
		{
			name: "single file upload",
			formData: multipartFormData{
				fields: map[string][]string{
					keyName: {"document"},
				},
				files: map[string][]fileData{
					bodyFile: {{filename: "test.txt", content: []byte("file content")}},
				},
			},
			bodyType: reflect.TypeFor[multipartFileStruct](),
			//nolint:thelper // inline test validation
			checkFunc: func(t *testing.T, result map[string]any) {
				body, ok := result[goFieldBody].(map[string]any)
				require.True(t, ok)
				assert.Equal(t, "document", body[keyName])

				file, ok := body[bodyFile].(io.ReadCloser)
				require.True(t, ok)
				defer file.Close() //nolint:errcheck // test cleanup

				content, err := io.ReadAll(file)
				require.NoError(t, err)
				assert.Equal(t, []byte("file content"), content)
			},
		},
		{
			name: "multiple files same field",
			formData: multipartFormData{
				files: map[string][]fileData{
					"files": {
						{filename: "file1.txt", content: []byte("content1")},
						{filename: "file2.txt", content: []byte("content2")},
					},
				},
			},
			bodyType: reflect.TypeFor[multipartMultiFileStruct](),
			//nolint:thelper // inline test validation
			checkFunc: func(t *testing.T, result map[string]any) {
				body, ok := result[goFieldBody].(map[string]any)
				require.True(t, ok)

				files, ok := body["files"].([]io.ReadCloser)
				require.True(t, ok)
				require.Len(t, files, 2)

				for i, f := range files {
					defer f.Close() //nolint:errcheck // test cleanup
					content, err := io.ReadAll(f)
					require.NoError(t, err)
					assert.Equal(t, []byte("content"+string(rune('1'+i))), content)
				}
			},
		},
		{
			name: "mixed text and file",
			formData: multipartFormData{
				fields: map[string][]string{
					"title": {"My Document"},
				},
				files: map[string][]fileData{
					"attachment": {{filename: "doc.pdf", content: []byte("PDF content")}},
				},
			},
			bodyType: reflect.TypeFor[multipartMixedStruct](),
			//nolint:thelper // inline test validation
			checkFunc: func(t *testing.T, result map[string]any) {
				body, ok := result[goFieldBody].(map[string]any)
				require.True(t, ok)
				assert.Equal(t, "My Document", body["title"])

				file, ok := body["attachment"].(io.ReadCloser)
				require.True(t, ok)
				defer file.Close() //nolint:errcheck // test cleanup

				content, err := io.ReadAll(file)
				require.NoError(t, err)
				assert.Equal(t, []byte("PDF content"), content)
			},
		},
		{
			name: "empty file",
			formData: multipartFormData{
				fields: map[string][]string{
					keyName: {"empty"},
				},
				files: map[string][]fileData{
					bodyFile: {{filename: "empty.txt", content: []byte{}}},
				},
			},
			bodyType: reflect.TypeFor[multipartFileStruct](),
			//nolint:thelper // inline test validation
			checkFunc: func(t *testing.T, result map[string]any) {
				body, ok := result[goFieldBody].(map[string]any)
				require.True(t, ok)

				file, ok := body[bodyFile].(io.ReadCloser)
				require.True(t, ok)
				defer file.Close() //nolint:errcheck // test cleanup

				content, err := io.ReadAll(file)
				require.NoError(t, err)
				assert.Empty(t, content)
			},
		},
		{
			name: "field in metadata but missing in form",
			formData: multipartFormData{
				fields: map[string][]string{
					goFieldName: {valueJohn},
					// email and Age are missing
				},
			},
			bodyType: reflect.TypeFor[multipartFormStruct](),
			want: map[string]any{
				goFieldBody: map[string]any{
					goFieldName: valueJohn,
				},
			},
		},
		{
			name: "unicode field values",
			formData: multipartFormData{
				fields: map[string][]string{
					goFieldName: {"日本語テスト"},
				},
			},
			bodyType: reflect.TypeFor[multipartFormStruct](),
			want: map[string]any{
				goFieldBody: map[string]any{
					goFieldName: "日本語テスト",
				},
			},
		},
	}

	decoder := newTestDecoder()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata := createBodyMetadata(BodyTypeMultipart, tt.bodyType)
			req := createMultipartRequest(tt.formData)

			result, err := decoder.Decode(req, nil, metadata)

			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)

			if tt.checkFunc != nil {
				tt.checkFunc(t, result)

				return
			}

			assert.Equal(t, tt.want, result)
		})
	}
}

func TestDecoder_DecodeBody_Multipart_Error(t *testing.T) {
	decoder := newTestDecoder()
	metadata := createBodyMetadata(BodyTypeMultipart, reflect.TypeFor[multipartFormStruct]())

	// Request without multipart content type
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader([]byte("not multipart")))
	req.Header.Set("Content-Type", "text/plain")

	_, err := decoder.Decode(req, nil, metadata)
	require.Error(t, err)
}

// urlEncodedFormStruct is used for URL-encoded form tests.
type urlEncodedFormStruct struct {
	Name    string // no tag - uses field name "Name"
	Email   string `schema:"email"`
	Age     string `schema:",required"` // no name, uses field name "Age"
	Ignored string `schema:"-"`
}

func TestDecoder_DecodeBody_URLEncodedForm(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		bodyType reflect.Type
		want     map[string]any
		wantErr  bool
	}{
		{
			name:     "single field",
			body:     "Name=John",
			bodyType: reflect.TypeFor[urlEncodedFormStruct](),
			want: map[string]any{
				goFieldBody: map[string]any{
					goFieldName: valueJohn,
				},
			},
		},
		{
			name:     "multiple fields",
			body:     "Name=John&email=john@example.com&Age=30",
			bodyType: reflect.TypeFor[urlEncodedFormStruct](),
			want: map[string]any{
				goFieldBody: map[string]any{
					goFieldName: valueJohn,
					keyEmail:    valueEmailJohn, // Tag name "email", not field name "Email"
					goFieldAge:  "30",
				},
			},
		},
		{
			name:     "empty form body",
			body:     "",
			bodyType: reflect.TypeFor[urlEncodedFormStruct](),
			want:     map[string]any{},
		},
		{
			name:     "field with empty value",
			body:     "Name=&email=test@example.com",
			bodyType: reflect.TypeFor[urlEncodedFormStruct](),
			want: map[string]any{
				goFieldBody: map[string]any{
					keyEmail: "test@example.com",
				},
			},
		},
		{
			name:     "unknown fields ignored",
			body:     "Name=John&unknown=ignored&other=also_ignored",
			bodyType: reflect.TypeFor[urlEncodedFormStruct](),
			want: map[string]any{
				goFieldBody: map[string]any{
					goFieldName: valueJohn,
				},
			},
		},
		{
			name:     "multiple values same key",
			body:     "Name=John&Name=Jane",
			bodyType: reflect.TypeFor[urlEncodedFormStruct](),
			want: map[string]any{
				goFieldBody: map[string]any{
					goFieldName: []any{valueJohn, valueJane},
				},
			},
		},
		{
			name:     "comma separated values",
			body:     "Name=John,Jane,Bob",
			bodyType: reflect.TypeFor[urlEncodedFormStruct](),
			want: map[string]any{
				goFieldBody: map[string]any{
					goFieldName: []any{valueJohn, valueJane, "Bob"},
				},
			},
		},
		{
			name:     "URL encoded special characters",
			body:     "Name=John%20Doe&email=john%2Bdoe%40example.com",
			bodyType: reflect.TypeFor[urlEncodedFormStruct](),
			want: map[string]any{
				goFieldBody: map[string]any{
					goFieldName: "John Doe",
					keyEmail:    "john+doe@example.com",
				},
			},
		},
		{
			name:     "unicode characters",
			body:     "Name=%E6%97%A5%E6%9C%AC%E8%AA%9E",
			bodyType: reflect.TypeFor[urlEncodedFormStruct](),
			want: map[string]any{
				goFieldBody: map[string]any{
					goFieldName: "日本語",
				},
			},
		},
		{
			name:     "field in metadata but missing in form",
			body:     "Name=John",
			bodyType: reflect.TypeFor[urlEncodedFormStruct](),
			want: map[string]any{
				goFieldBody: map[string]any{
					goFieldName: valueJohn,
				},
			},
		},
		{
			name:     "pointer to struct body type",
			body:     "Name=John&email=john@example.com",
			bodyType: reflect.TypeFor[*urlEncodedFormStruct](),
			want: map[string]any{
				goFieldBody: map[string]any{
					goFieldName: valueJohn,
					keyEmail:    valueEmailJohn, // Tag name "email", not field name "Email"
				},
			},
		},
	}

	decoder := newTestDecoder()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata := createBodyMetadata(BodyTypeStructured, tt.bodyType)
			req := createBodyRequest("application/x-www-form-urlencoded", []byte(tt.body))

			result, err := decoder.Decode(req, nil, metadata)

			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}
