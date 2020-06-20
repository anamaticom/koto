// Code generated by protoc-gen-twirp v5.12.0, DO NOT EDIT.
// source: user.proto

package rpc

import bytes "bytes"
import strings "strings"
import context "context"
import fmt "fmt"
import ioutil "io/ioutil"
import http "net/http"
import strconv "strconv"

import jsonpb "github.com/golang/protobuf/jsonpb"
import proto "github.com/golang/protobuf/proto"
import twirp "github.com/twitchtv/twirp"
import ctxsetters "github.com/twitchtv/twirp/ctxsetters"

// =====================
// UserService Interface
// =====================

type UserService interface {
	Friends(context.Context, *Empty) (*UserFriendsResponse, error)
}

// ===========================
// UserService Protobuf Client
// ===========================

type userServiceProtobufClient struct {
	client HTTPClient
	urls   [1]string
	opts   twirp.ClientOptions
}

// NewUserServiceProtobufClient creates a Protobuf client that implements the UserService interface.
// It communicates using Protobuf and can be configured with a custom HTTPClient.
func NewUserServiceProtobufClient(addr string, client HTTPClient, opts ...twirp.ClientOption) UserService {
	if c, ok := client.(*http.Client); ok {
		client = withoutRedirects(c)
	}

	clientOpts := twirp.ClientOptions{}
	for _, o := range opts {
		o(&clientOpts)
	}

	prefix := urlBase(addr) + UserServicePathPrefix
	urls := [1]string{
		prefix + "Friends",
	}

	return &userServiceProtobufClient{
		client: client,
		urls:   urls,
		opts:   clientOpts,
	}
}

func (c *userServiceProtobufClient) Friends(ctx context.Context, in *Empty) (*UserFriendsResponse, error) {
	ctx = ctxsetters.WithPackageName(ctx, "rpc")
	ctx = ctxsetters.WithServiceName(ctx, "UserService")
	ctx = ctxsetters.WithMethodName(ctx, "Friends")
	out := new(UserFriendsResponse)
	err := doProtobufRequest(ctx, c.client, c.opts.Hooks, c.urls[0], in, out)
	if err != nil {
		twerr, ok := err.(twirp.Error)
		if !ok {
			twerr = twirp.InternalErrorWith(err)
		}
		callClientError(ctx, c.opts.Hooks, twerr)
		return nil, err
	}

	callClientResponseReceived(ctx, c.opts.Hooks)

	return out, nil
}

// =======================
// UserService JSON Client
// =======================

type userServiceJSONClient struct {
	client HTTPClient
	urls   [1]string
	opts   twirp.ClientOptions
}

// NewUserServiceJSONClient creates a JSON client that implements the UserService interface.
// It communicates using JSON and can be configured with a custom HTTPClient.
func NewUserServiceJSONClient(addr string, client HTTPClient, opts ...twirp.ClientOption) UserService {
	if c, ok := client.(*http.Client); ok {
		client = withoutRedirects(c)
	}

	clientOpts := twirp.ClientOptions{}
	for _, o := range opts {
		o(&clientOpts)
	}

	prefix := urlBase(addr) + UserServicePathPrefix
	urls := [1]string{
		prefix + "Friends",
	}

	return &userServiceJSONClient{
		client: client,
		urls:   urls,
		opts:   clientOpts,
	}
}

func (c *userServiceJSONClient) Friends(ctx context.Context, in *Empty) (*UserFriendsResponse, error) {
	ctx = ctxsetters.WithPackageName(ctx, "rpc")
	ctx = ctxsetters.WithServiceName(ctx, "UserService")
	ctx = ctxsetters.WithMethodName(ctx, "Friends")
	out := new(UserFriendsResponse)
	err := doJSONRequest(ctx, c.client, c.opts.Hooks, c.urls[0], in, out)
	if err != nil {
		twerr, ok := err.(twirp.Error)
		if !ok {
			twerr = twirp.InternalErrorWith(err)
		}
		callClientError(ctx, c.opts.Hooks, twerr)
		return nil, err
	}

	callClientResponseReceived(ctx, c.opts.Hooks)

	return out, nil
}

// ==========================
// UserService Server Handler
// ==========================

type userServiceServer struct {
	UserService
	hooks *twirp.ServerHooks
}

func NewUserServiceServer(svc UserService, hooks *twirp.ServerHooks) TwirpServer {
	return &userServiceServer{
		UserService: svc,
		hooks:       hooks,
	}
}

// writeError writes an HTTP response with a valid Twirp error format, and triggers hooks.
// If err is not a twirp.Error, it will get wrapped with twirp.InternalErrorWith(err)
func (s *userServiceServer) writeError(ctx context.Context, resp http.ResponseWriter, err error) {
	writeError(ctx, resp, err, s.hooks)
}

// UserServicePathPrefix is used for all URL paths on a twirp UserService server.
// Requests are always: POST UserServicePathPrefix/method
// It can be used in an HTTP mux to route twirp requests along with non-twirp requests on other routes.
const UserServicePathPrefix = "/rpc.UserService/"

func (s *userServiceServer) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	ctx = ctxsetters.WithPackageName(ctx, "rpc")
	ctx = ctxsetters.WithServiceName(ctx, "UserService")
	ctx = ctxsetters.WithResponseWriter(ctx, resp)

	var err error
	ctx, err = callRequestReceived(ctx, s.hooks)
	if err != nil {
		s.writeError(ctx, resp, err)
		return
	}

	if req.Method != "POST" {
		msg := fmt.Sprintf("unsupported method %q (only POST is allowed)", req.Method)
		err = badRouteError(msg, req.Method, req.URL.Path)
		s.writeError(ctx, resp, err)
		return
	}

	switch req.URL.Path {
	case "/rpc.UserService/Friends":
		s.serveFriends(ctx, resp, req)
		return
	default:
		msg := fmt.Sprintf("no handler for path %q", req.URL.Path)
		err = badRouteError(msg, req.Method, req.URL.Path)
		s.writeError(ctx, resp, err)
		return
	}
}

func (s *userServiceServer) serveFriends(ctx context.Context, resp http.ResponseWriter, req *http.Request) {
	header := req.Header.Get("Content-Type")
	i := strings.Index(header, ";")
	if i == -1 {
		i = len(header)
	}
	switch strings.TrimSpace(strings.ToLower(header[:i])) {
	case "application/json":
		s.serveFriendsJSON(ctx, resp, req)
	case "application/protobuf":
		s.serveFriendsProtobuf(ctx, resp, req)
	default:
		msg := fmt.Sprintf("unexpected Content-Type: %q", req.Header.Get("Content-Type"))
		twerr := badRouteError(msg, req.Method, req.URL.Path)
		s.writeError(ctx, resp, twerr)
	}
}

func (s *userServiceServer) serveFriendsJSON(ctx context.Context, resp http.ResponseWriter, req *http.Request) {
	var err error
	ctx = ctxsetters.WithMethodName(ctx, "Friends")
	ctx, err = callRequestRouted(ctx, s.hooks)
	if err != nil {
		s.writeError(ctx, resp, err)
		return
	}

	reqContent := new(Empty)
	unmarshaler := jsonpb.Unmarshaler{AllowUnknownFields: true}
	if err = unmarshaler.Unmarshal(req.Body, reqContent); err != nil {
		s.writeError(ctx, resp, malformedRequestError("the json request could not be decoded"))
		return
	}

	// Call service method
	var respContent *UserFriendsResponse
	func() {
		defer ensurePanicResponses(ctx, resp, s.hooks)
		respContent, err = s.UserService.Friends(ctx, reqContent)
	}()

	if err != nil {
		s.writeError(ctx, resp, err)
		return
	}
	if respContent == nil {
		s.writeError(ctx, resp, twirp.InternalError("received a nil *UserFriendsResponse and nil error while calling Friends. nil responses are not supported"))
		return
	}

	ctx = callResponsePrepared(ctx, s.hooks)

	var buf bytes.Buffer
	marshaler := &jsonpb.Marshaler{OrigName: true}
	if err = marshaler.Marshal(&buf, respContent); err != nil {
		s.writeError(ctx, resp, wrapInternal(err, "failed to marshal json response"))
		return
	}

	ctx = ctxsetters.WithStatusCode(ctx, http.StatusOK)
	respBytes := buf.Bytes()
	resp.Header().Set("Content-Type", "application/json")
	resp.Header().Set("Content-Length", strconv.Itoa(len(respBytes)))
	resp.WriteHeader(http.StatusOK)

	if n, err := resp.Write(respBytes); err != nil {
		msg := fmt.Sprintf("failed to write response, %d of %d bytes written: %s", n, len(respBytes), err.Error())
		twerr := twirp.NewError(twirp.Unknown, msg)
		callError(ctx, s.hooks, twerr)
	}
	callResponseSent(ctx, s.hooks)
}

func (s *userServiceServer) serveFriendsProtobuf(ctx context.Context, resp http.ResponseWriter, req *http.Request) {
	var err error
	ctx = ctxsetters.WithMethodName(ctx, "Friends")
	ctx, err = callRequestRouted(ctx, s.hooks)
	if err != nil {
		s.writeError(ctx, resp, err)
		return
	}

	buf, err := ioutil.ReadAll(req.Body)
	if err != nil {
		s.writeError(ctx, resp, wrapInternal(err, "failed to read request body"))
		return
	}
	reqContent := new(Empty)
	if err = proto.Unmarshal(buf, reqContent); err != nil {
		s.writeError(ctx, resp, malformedRequestError("the protobuf request could not be decoded"))
		return
	}

	// Call service method
	var respContent *UserFriendsResponse
	func() {
		defer ensurePanicResponses(ctx, resp, s.hooks)
		respContent, err = s.UserService.Friends(ctx, reqContent)
	}()

	if err != nil {
		s.writeError(ctx, resp, err)
		return
	}
	if respContent == nil {
		s.writeError(ctx, resp, twirp.InternalError("received a nil *UserFriendsResponse and nil error while calling Friends. nil responses are not supported"))
		return
	}

	ctx = callResponsePrepared(ctx, s.hooks)

	respBytes, err := proto.Marshal(respContent)
	if err != nil {
		s.writeError(ctx, resp, wrapInternal(err, "failed to marshal proto response"))
		return
	}

	ctx = ctxsetters.WithStatusCode(ctx, http.StatusOK)
	resp.Header().Set("Content-Type", "application/protobuf")
	resp.Header().Set("Content-Length", strconv.Itoa(len(respBytes)))
	resp.WriteHeader(http.StatusOK)
	if n, err := resp.Write(respBytes); err != nil {
		msg := fmt.Sprintf("failed to write response, %d of %d bytes written: %s", n, len(respBytes), err.Error())
		twerr := twirp.NewError(twirp.Unknown, msg)
		callError(ctx, s.hooks, twerr)
	}
	callResponseSent(ctx, s.hooks)
}

func (s *userServiceServer) ServiceDescriptor() ([]byte, int) {
	return twirpFileDescriptor5, 0
}

func (s *userServiceServer) ProtocGenTwirpVersion() string {
	return "v5.12.0"
}

func (s *userServiceServer) PathPrefix() string {
	return UserServicePathPrefix
}

var twirpFileDescriptor5 = []byte{
	// 147 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x2a, 0x2d, 0x4e, 0x2d,
	0xd2, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x2e, 0x2a, 0x48, 0x96, 0xe2, 0xce, 0xcd, 0x4f,
	0x49, 0xcd, 0x81, 0x88, 0x28, 0x59, 0x71, 0x09, 0x87, 0x16, 0xa7, 0x16, 0xb9, 0x15, 0x65, 0xa6,
	0xe6, 0xa5, 0x14, 0x07, 0xa5, 0x16, 0x17, 0xe4, 0xe7, 0x15, 0xa7, 0x0a, 0x29, 0x73, 0xb1, 0xa7,
	0x41, 0x84, 0x24, 0x18, 0x15, 0x98, 0x35, 0xb8, 0x8d, 0x38, 0xf5, 0x8a, 0x0a, 0x92, 0xf5, 0x40,
	0x4a, 0x83, 0x60, 0x32, 0x46, 0x76, 0x5c, 0xdc, 0x20, 0x81, 0xe0, 0xd4, 0xa2, 0xb2, 0xcc, 0xe4,
	0x54, 0x21, 0x7d, 0x2e, 0x76, 0xa8, 0x31, 0x42, 0x5c, 0x60, 0xd5, 0xae, 0xb9, 0x05, 0x25, 0x95,
	0x52, 0x12, 0x70, 0x9d, 0x68, 0x96, 0x38, 0x71, 0x44, 0xb1, 0xe9, 0xe9, 0xe9, 0x17, 0x15, 0x24,
	0x27, 0xb1, 0x81, 0x1d, 0x63, 0x0c, 0x08, 0x00, 0x00, 0xff, 0xff, 0x1e, 0xf2, 0xb5, 0x0e, 0xac,
	0x00, 0x00, 0x00,
}