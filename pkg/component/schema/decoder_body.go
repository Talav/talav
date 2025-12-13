package schema

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"reflect"
)

// decodeBody decodes the HTTP request body based on content type.
func (d *defaultDecoder) decodeBody(request *http.Request, metadata *StructMetadata) (map[string]any, error) {
	bodyField := metadata.BodyField()
	if bodyField == nil {
		return make(map[string]any), nil
	}

	bodyContentType := newBodyContentType(request.Header.Get("Content-Type"), bodyField.BodyType)

	// Multipart needs raw request for ParseMultipartForm - handle before reading body
	if bodyContentType.isMultipart() {
		return d.decodeMultipartBody(request, bodyField)
	}

	// Read body for other content types
	bodyBytes, err := io.ReadAll(request.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	if len(bodyBytes) == 0 {
		return make(map[string]any), nil
	}

	if bodyContentType.isForm() {
		return d.decodeURLEncodedForm(bodyBytes, bodyField)
	}

	if bodyContentType.isFile() {
		return d.decodeFileBody(bodyBytes, bodyField)
	}

	if bodyContentType.isXML() {
		return d.decodeXMLBody(bodyBytes, bodyField)
	}

	// try JSON as a fallback
	return d.decodeJSONBody(bodyBytes, bodyField)
}

// decodeXMLBody decodes XML body content.
func (d *defaultDecoder) decodeXMLBody(bodyBytes []byte, bodyField *FieldMetadata) (map[string]any, error) {
	var parsed any
	if err := xml.Unmarshal(bodyBytes, &parsed); err != nil {
		return nil, fmt.Errorf("failed to unmarshal XML: %w", err)
	}

	return map[string]any{bodyField.MapKey: parsed}, nil
}

// decodeJSONBody decodes JSON body content.
func (d *defaultDecoder) decodeJSONBody(bodyBytes []byte, bodyField *FieldMetadata) (map[string]any, error) {
	var parsed any
	if err := json.Unmarshal(bodyBytes, &parsed); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return map[string]any{bodyField.MapKey: parsed}, nil
}

// decodeURLEncodedForm decodes URL-encoded form body content.
func (d *defaultDecoder) decodeURLEncodedForm(bodyBytes []byte, bodyField *FieldMetadata) (map[string]any, error) {
	// Use decodeFormStyle to parse form data (same logic as query parameters)
	decodedMap, err := d.decodeFormStyle(string(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to parse form: %w", err)
	}

	// Filter map to only include fields from the body struct (using schema tags)
	structType := bodyField.Type
	if structType.Kind() == kindPtr {
		structType = structType.Elem()
	}

	// Get metadata for the body struct to filter by schema tags
	metadata, err := d.structMetadataCache.getStructMetadata(structType)
	if err != nil {
		return nil, fmt.Errorf("failed to get struct metadata: %w", err)
	}

	// Filter decoded map to only include fields that exist in the struct
	filteredMap := make(map[string]any)
	for _, fieldMeta := range metadata.Fields {
		// Skip body fields (shouldn't have nested body fields)
		if fieldMeta.IsBody {
			continue
		}

		// Use ParamName (from schema tag) to look up in decoded map
		if value, exists := decodedMap[fieldMeta.ParamName]; exists {
			filteredMap[fieldMeta.MapKey] = value
		}
	}

	return map[string]any{bodyField.MapKey: filteredMap}, nil
}

// decodeMultipartBody decodes multipart form body content.
func (d *defaultDecoder) decodeMultipartBody(r *http.Request, bodyField *FieldMetadata) (map[string]any, error) {
	// Parse multipart form
	if err := r.ParseMultipartForm(32 << 20); err != nil { // 32MB max
		return nil, fmt.Errorf("failed to parse multipart form: %w", err)
	}

	form := r.MultipartForm
	if form == nil {
		return nil, fmt.Errorf("not a multipart form")
	}

	// Get the struct type
	structType := bodyField.Type
	if structType.Kind() == kindPtr {
		structType = structType.Elem()
	}

	// Get metadata for the body struct to use schema tags
	metadata, err := d.structMetadataCache.getStructMetadata(structType)
	if err != nil {
		return nil, fmt.Errorf("failed to get struct metadata: %w", err)
	}

	fieldMap := make(map[string]any)

	// Process each field using cached metadata
	for _, fieldMeta := range metadata.Fields {
		if err := d.processMultipartField(form, fieldMeta, fieldMap); err != nil {
			return nil, err
		}
	}

	return map[string]any{bodyField.MapKey: fieldMap}, nil
}

// decodeFileBody decodes raw file body content.
func (d *defaultDecoder) decodeFileBody(bodyBytes []byte, bodyField *FieldMetadata) (map[string]any, error) {
	// For file upload, read the body and store as []byte in map
	result := make(map[string]any)
	result[bodyField.MapKey] = bodyBytes

	return result, nil
}

// processMultipartField processes a single multipart form field (file or regular).
func (d *defaultDecoder) processMultipartField(form *multipart.Form, fieldMeta FieldMetadata, result map[string]any) error {
	// Detect file fields by type
	if d.isFileField(fieldMeta.Type) {
		// File field: open files and return as io.ReadCloser (standard format)
		fileHeaders := form.File[fieldMeta.ParamName]
		if len(fileHeaders) > 0 {
			fileReaders, err := d.openMultipartFiles(fileHeaders, fieldMeta.Type)
			if err != nil {
				return fmt.Errorf("failed to open file field %s: %w", fieldMeta.ParamName, err)
			}
			result[fieldMeta.MapKey] = fileReaders
		}

		return nil
	}

	// Regular form field: use schema tag name
	values := form.Value[fieldMeta.ParamName]
	if len(values) > 0 {
		if len(values) == 1 {
			result[fieldMeta.MapKey] = values[0]
		} else {
			result[fieldMeta.MapKey] = stringSliceToAny(values)
		}
	}

	return nil
}

// isFileField checks if a type represents a file field.
func (d *defaultDecoder) isFileField(typ reflect.Type) bool {
	// Check for io.ReadCloser
	if typ.Implements(reflect.TypeOf((*io.ReadCloser)(nil)).Elem()) {
		return true
	}

	// Check for []byte
	if typ.Kind() == kindSlice && typ.Elem().Kind() == reflect.Uint8 {
		return true
	}

	// Check for []io.ReadCloser (slice of files)
	if typ.Kind() == kindSlice {
		elemType := typ.Elem()
		if elemType.Implements(reflect.TypeOf((*io.ReadCloser)(nil)).Elem()) {
			return true
		}
	}

	return false
}

// openMultipartFiles opens multipart file headers and returns io.ReadCloser(s).
// Returns io.ReadCloser for single file fields, []io.ReadCloser for multiple file fields.
func (d *defaultDecoder) openMultipartFiles(fileHeaders []*multipart.FileHeader, fieldType reflect.Type) (any, error) {
	// Check if field is a slice
	if fieldType.Kind() == kindSlice {
		// Multiple files - return []io.ReadCloser
		readers := make([]io.ReadCloser, len(fileHeaders))
		for i, fh := range fileHeaders {
			file, err := fh.Open()
			if err != nil {
				// Close already opened files on error
				for j := 0; j < i; j++ {
					_ = readers[j].Close()
				}

				return nil, fmt.Errorf("failed to open file[%d]: %w", i, err)
			}
			readers[i] = file
		}

		return readers, nil
	}

	file, err := fileHeaders[0].Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return file, nil
}
