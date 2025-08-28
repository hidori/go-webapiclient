package webapiclient

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientImpl_New(t *testing.T) {
	t.Parallel()

	mockDoFunc := func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader([]byte("ok"))),
		}, nil
	}

	type args struct {
		do      DoFunc
		baseURL string
	}
	tests := []struct {
		name string
		args args
		want *client
	}{
		{
			name: "success: valid parameters",
			args: args{
				do:      mockDoFunc,
				baseURL: "http://example.com",
			},
			want: &client{
				do:      mockDoFunc,
				baseURL: "http://example.com",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := NewClient(tt.args.do, tt.args.baseURL)
			require.NotNil(t, got)
			assertEqual(t, tt.want, got, cmp.AllowUnexported(client{}), cmpopts.IgnoreFields(client{}, "do"))

			clientImpl := got.(*client)
			assert.NotNil(t, clientImpl.do)
		})
	}
}

func assertEqual(t *testing.T, want any, got any, options ...cmp.Option) {
	t.Helper()
	if diff := cmp.Diff(want, got, options...); diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}
}

func TestClientImpl_Do(t *testing.T) {
	t.Parallel()

	type fields struct {
		baseURL string
		do      func(req *http.Request) (*http.Response, error)
	}
	type args struct {
		ctx     context.Context
		request *Request
		edit    EditRequestFunc
	}
	type want struct {
		err    bool
		status int
		body   []byte
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "success: GET request",
			fields: fields{
				baseURL: "http://example.com",
				do: func(req *http.Request) (*http.Response, error) {
					assert.Equal(t, http.MethodGet, req.Method)
					assert.Equal(t, "http://example.com/test", req.URL.String())
					assert.Equal(t, "application/json", req.Header.Get("Accept"))
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewReader([]byte("test response"))),
					}, nil
				},
			},
			args: args{
				ctx: context.Background(),
				request: &Request{
					Method: http.MethodGet,
					Path:   "/test",
					Headers: map[string][]string{
						"Accept": {"application/json"},
					},
				},
				edit: nil,
			},
			want: want{
				err:    false,
				status: http.StatusOK,
				body:   []byte("test response"),
			},
		},
		{
			name: "success: POST request",
			fields: fields{
				baseURL: "http://example.com",
				do: func(req *http.Request) (*http.Response, error) {
					assert.Equal(t, http.MethodPost, req.Method)
					assert.Equal(t, "http://example.com/test", req.URL.String())
					assert.Equal(t, "application/json", req.Header.Get("Content-Type"))

					body, err := io.ReadAll(req.Body)
					require.NoError(t, err)
					assert.Equal(t, []byte(`{"test":"data"}`), body)

					return &http.Response{
						StatusCode: http.StatusCreated,
						Body:       io.NopCloser(bytes.NewReader([]byte("created response"))),
					}, nil
				},
			},
			args: args{
				ctx: context.Background(),
				request: &Request{
					Method: http.MethodPost,
					Path:   "/test",
					Headers: map[string][]string{
						"Content-Type": {"application/json"},
					},
					Body: bytes.NewReader([]byte(`{"test":"data"}`)),
				},
				edit: nil,
			},
			want: want{
				err:    false,
				status: http.StatusCreated,
				body:   []byte("created response"),
			},
		},
		{
			name: "success: PUT request",
			fields: fields{
				baseURL: "http://example.com",
				do: func(req *http.Request) (*http.Response, error) {
					assert.Equal(t, http.MethodPut, req.Method)
					assert.Equal(t, "http://example.com/test", req.URL.String())
					assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewReader([]byte("updated response"))),
					}, nil
				},
			},
			args: args{
				ctx: context.Background(),
				request: &Request{
					Method: http.MethodPut,
					Path:   "/test",
					Headers: map[string][]string{
						"Content-Type": {"application/json"},
					},
				},
				edit: nil,
			},
			want: want{
				err:    false,
				status: http.StatusOK,
				body:   []byte("updated response"),
			},
		},
		{
			name: "success: DELETE request",
			fields: fields{
				baseURL: "http://example.com",
				do: func(req *http.Request) (*http.Response, error) {
					assert.Equal(t, http.MethodDelete, req.Method)
					assert.Equal(t, "http://example.com/test", req.URL.String())
					return &http.Response{
						StatusCode: http.StatusNoContent,
						Body:       io.NopCloser(bytes.NewReader([]byte{})),
					}, nil
				},
			},
			args: args{
				ctx: context.Background(),
				request: &Request{
					Method: http.MethodDelete,
					Path:   "/test",
				},
				edit: nil,
			},
			want: want{
				err:    false,
				status: http.StatusNoContent,
				body:   []byte{},
			},
		},
		{
			name: "success: edit request",
			fields: fields{
				baseURL: "http://example.com",
				do: func(req *http.Request) (*http.Response, error) {
					assert.Equal(t, http.MethodPost, req.Method)
					assert.Equal(t, "http://example.com/test", req.URL.String())
					assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
					assert.Equal(t, "custom-value", req.Header.Get("X-Custom-Header"))
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewReader([]byte("edited request response"))),
					}, nil
				},
			},
			args: args{
				ctx: context.Background(),
				request: &Request{
					Method: http.MethodPost,
					Path:   "/test",
					Headers: map[string][]string{
						"Content-Type": {"application/json"},
					},
				},
				edit: func(req *http.Request) error {
					req.Header.Set("X-Custom-Header", "custom-value")
					return nil
				},
			},
			want: want{
				err:    false,
				status: http.StatusOK,
				body:   []byte("edited request response"),
			},
		},
		{
			name: "failure: unexpected status code",
			fields: fields{
				baseURL: "http://example.com",
				do: func(req *http.Request) (*http.Response, error) {
					assert.Equal(t, http.MethodGet, req.Method)
					assert.Equal(t, "http://example.com/test", req.URL.String())
					assert.Equal(t, "application/json", req.Header.Get("Accept"))
					return &http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       io.NopCloser(bytes.NewReader([]byte("internal server error"))),
					}, nil
				},
			},
			args: args{
				ctx: context.Background(),
				request: &Request{
					Method: http.MethodGet,
					Path:   "/test",
					Headers: map[string][]string{
						"Accept": {"application/json"},
					},
					ExpectedStatusCodes: []int{http.StatusOK},
				},
				edit: nil,
			},
			want: want{
				err: true,
			},
		},
		{
			name: "failure: unexpected content-type",
			fields: fields{
				baseURL: "http://example.com",
				do: func(req *http.Request) (*http.Response, error) {
					assert.Equal(t, http.MethodGet, req.Method)
					assert.Equal(t, "http://example.com/test", req.URL.String())
					return &http.Response{
						StatusCode: http.StatusOK,
						Header:     http.Header{"Content-Type": []string{"text/plain"}},
						Body:       io.NopCloser(bytes.NewReader([]byte("unexpected content"))),
					}, nil
				},
			},
			args: args{
				ctx: context.Background(),
				request: &Request{
					Method:               http.MethodGet,
					Path:                 "/test",
					ExpectedContentTypes: []string{"application/json"},
				},
				edit: nil,
			},
			want: want{
				err: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := NewClient(tt.fields.do, tt.fields.baseURL)

			got, err := client.Do(tt.args.ctx, tt.args.request, tt.args.edit)
			if tt.want.err {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			defer func() {
				_ = got.Body.Close()
			}()
			assert.Equal(t, tt.want.status, got.StatusCode)

			actualBody, err := io.ReadAll(got.Body)
			require.NoError(t, err)
			assert.Equal(t, tt.want.body, actualBody)
		})
	}
}
