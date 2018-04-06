//
// DISCLAIMER
//
// Copyright 2017 ArangoDB GmbH, Cologne, Germany
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Copyright holder is ArangoDB GmbH, Cologne, Germany
//
// Author Ewout Prangsma
//

package driver

import (
	"context"
	"encoding/json"
	"errors"

	velocypack "github.com/arangodb/go-velocypack"
)

// Connection is a connenction to a database server using a specific protocol.
type Connection interface {
	// NewRequest creates a new request with given method and path.
	NewRequest(method, path string) (Request, error)

	// Do performs a given request, returning its response.
	Do(ctx context.Context, req Request) (Response, error)

	// DoDecode performs same as Do() but deserialize Body in `obj`
	DoDecode(ctx context.Context, req Request, obj interface{}) (Response, error)

	// Unmarshal unmarshals the given raw object into the given result interface.
	Unmarshal(data RawObject, result interface{}) error

	// Endpoints returns the endpoints used by this connection.
	Endpoints() []string

	// UpdateEndpoints reconfigures the connection to use the given endpoints.
	UpdateEndpoints(endpoints []string) error

	// Configure the authentication used for this connection.
	SetAuthentication(Authentication) (Connection, error)

	// Protocols returns all protocols used by this connection.
	Protocols() ProtocolSet
}

// Request represents the input to a request on the server.
type Request interface {
	// SetQuery sets a single query argument of the request.
	// Any existing query argument with the same key is overwritten.
	SetQuery(key, value string) Request
	// SetBody sets the content of the request.
	// The protocol of the connection determines what kinds of marshalling is taking place.
	// When multiple bodies are given, they are merged, with fields in the first document prevailing.
	SetBody(body ...interface{}) (Request, error)
	// SetBodyArray sets the content of the request as an array.
	// If the given mergeArray is not nil, its elements are merged with the elements in the body array (mergeArray data overrides bodyArray data).
	// The merge is NOT recursive.
	// The protocol of the connection determines what kinds of marshalling is taking place.
	SetBodyArray(bodyArray interface{}, mergeArray []map[string]interface{}) (Request, error)
	// SetBodyImportArray sets the content of the request as an array formatted for importing documents.
	// The protocol of the connection determines what kinds of marshalling is taking place.
	SetBodyImportArray(bodyArray interface{}) (Request, error)
	// SetHeader sets a single header arguments of the request.
	// Any existing header argument with the same key is overwritten.
	SetHeader(key, value string) Request
	// Written returns true as soon as this request has been written completely to the network.
	// This does not guarantee that the server has received or processed the request.
	Written() bool
	// Clone creates a new request containing the same data as this request
	Clone() Request
}

// Response represents the response from the server on a given request.
type Response interface {
	// StatusCode returns an HTTP compatible status code of the response.
	StatusCode() int
	// Endpoint returns the endpoint that handled the request.
	Endpoint() string
	// CheckStatus checks if the status of the response equals to one of the given status codes.
	// If so, nil is returned.
	// If not, an attempt is made to parse an error response in the body and an error is returned.
	CheckStatus(validStatusCodes ...int) error
	// Header returns the value of a response header with given key.
	// If no such header is found, an empty string is returned.
	// On nested Response's, this function will always return an empty string.
	Header(key string) string
	// ParseBody performs protocol specific unmarshalling of the response data into the given result.
	// If the given field is non-empty, the contents of that field will be parsed into the given result.
	// This can only be used for requests that return a single object.
	ParseBody(field string, result interface{}) error
	// ParseArrayBody performs protocol specific unmarshalling of the response array data into individual response objects.
	// This can only be used for requests that return an array of objects.
	ParseArrayBody() ([]Response, error)
}

// RawObject is a raw encoded object.
// Connection implementations must be able to unmarshal *RawObject into Go objects.
type RawObject []byte

// MarshalJSON returns *r as the JSON encoding of r.
func (r *RawObject) MarshalJSON() ([]byte, error) {
	return *r, nil
}

// UnmarshalJSON sets *r to a copy of data.
func (r *RawObject) UnmarshalJSON(data []byte) error {
	if r == nil {
		return errors.New("RawObject: UnmarshalJSON on nil pointer")
	}
	*r = append((*r)[0:0], data...)
	return nil
}

// Ensure RawObject implements json.Marshaler & json.Unmarshaler
var _ json.Marshaler = (*RawObject)(nil)
var _ json.Unmarshaler = (*RawObject)(nil)

// MarshalVPack returns m as the Velocypack encoding of m.
func (r RawObject) MarshalVPack() (velocypack.Slice, error) {
	if r == nil {
		return velocypack.NullSlice(), nil
	}
	return velocypack.Slice(r), nil
}

// UnmarshalVPack sets *m to a copy of data.
func (r *RawObject) UnmarshalVPack(data velocypack.Slice) error {
	if r == nil {
		return errors.New("velocypack.RawSlice: UnmarshalVPack on nil pointer")
	}
	*r = append((*r)[0:0], data...)
	return nil
}

var _ velocypack.Marshaler = (*RawObject)(nil)
var _ velocypack.Unmarshaler = (*RawObject)(nil)
