// Copyright © 2020 Banzai Cloud
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ale

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
	"time"

	"emperror.dev/errors"
	al_proto "github.com/envoyproxy/go-control-plane/envoy/data/accesslog/v2"
	als_proto_cp "github.com/envoyproxy/go-control-plane/envoy/service/accesslog/v2"
	"github.com/golang/protobuf/ptypes/duration"
)

func New(le *al_proto.HTTPAccessLogEntry) (*HTTPAccessLogEntry, error) {
	cp := le.GetCommonProperties()

	e := &HTTPAccessLogEntry{
		Direction:       "outbound",
		UpstreamCluster: le.GetCommonProperties().GetUpstreamCluster(),
		ProtocolVersion: le.GetProtocolVersion().String(),

		entry: le,
	}

	if strings.HasPrefix(cp.GetUpstreamCluster(), "inbound") {
		e.Direction = "inbound"
	}

	startTime := time.Unix(le.GetCommonProperties().GetStartTime().GetSeconds(), int64(le.GetCommonProperties().GetStartTime().GetNanos()))
	e.startTime = &startTime
	e.StartTime = startTime.String()

	e.Latency = e.getDurationFromProto(cp.GetTimeToLastDownstreamTxByte())
	e.setDurations()
	e.setSource()
	e.setSourceMetadata()
	e.setDestination()
	e.setDestinationMetadata()
	e.setRequest()
	e.setResponse()
	e.setAuthinfo()

	return e, nil
}

func (e *HTTPAccessLogEntry) GetUniqueKey(byWorkload bool) string {
	if byWorkload {
		return fmt.Sprintf("%s|%s|%s|%s|%s", e.Source.Workload, e.Destination.Workload, e.Request.Method, e.Request.Path, e.Request.Authority)
	}
	return fmt.Sprintf("%s|%s|%s|%s|%s", e.Source, e.Destination, e.Request.Method, e.Request.Path, e.Request.Authority)
}

func (e *HTTPAccessLogEntry) FormatedStartTime(format ...string) string {
	if len(format) != 1 || format[0] == "" {
		return e.startTime.Format(time.RFC3339)
	}

	return e.startTime.Format(format[0])
}

func (e *HTTPAccessLogEntry) FormattedResponseFlags() string {
	return strings.Join(e.Response.Flags, ",")
}

func (e *HTTPAccessLogEntry) ResponseBytes() uint64 {
	return e.Response.HeaderBytes + e.Response.BodyBytes
}

func (e *HTTPAccessLogEntry) RequestBytes() uint64 {
	return e.Request.HeaderBytes + e.Request.BodyBytes
}

func (e *HTTPAccessLogEntry) LatencyInMiliseconds() int64 {
	return e.Latency.Milliseconds()
}

func (e *HTTPAccessLogEntry) FormattedString(tpl *template.Template) string {
	var buf bytes.Buffer
	tpl.Execute(&buf, e)

	return buf.String()
}

func (e *HTTPAccessLogEntry) String() string {
	tpl, _ := template.New("format").Parse(`[{{.FormatedStartTime}}] "{{.Request.Method}} {{.Request.Path}} {{.ProtocolVersion}}" {{.Response.StatusCode}} {{.FormattedResponseFlags}} {{.ResponseBytes}} {{.RequestBytes}} {{.LatencyInMiliseconds}} "{{.Request.ForwardedFor}}" "{{.Request.UserAgent}}" "{{.Request.ID}}" "{{.Request.Authority}}" "{{.Destination.Address}}"`)

	return e.FormattedString(tpl)
}

func (e *HTTPAccessLogEntry) SetReporter(id *als_proto_cp.StreamAccessLogsMessage_Identifier) {
	node := id.GetNode()
	if node == nil {
		return
	}

	e.Reporter = &Reporter{
		ID: node.GetId(),
	}

	e.Reporter.SetAttributes(getStructValues(node.GetMetadata()))
}

func (e *HTTPAccessLogEntry) getDurationFromProto(d *duration.Duration) *time.Duration {
	t := time.Duration(d.GetNanos()) + (time.Duration(d.GetSeconds()) * time.Second)

	return &t
}

func (e *HTTPAccessLogEntry) setDurations() {
	cp := e.entry.GetCommonProperties()
	e.Durations = &RequestDurations{
		TimeToLastRxByte:            e.getDurationFromProto(cp.GetTimeToLastRxByte()),
		TimeToFirstUpstreamTxByte:   e.getDurationFromProto(cp.GetTimeToFirstUpstreamTxByte()),
		TimeToLastUpstreamTxByte:    e.getDurationFromProto(cp.GetTimeToLastUpstreamTxByte()),
		TimeToFirstUpstreamRxByte:   e.getDurationFromProto(cp.GetTimeToFirstUpstreamRxByte()),
		TimeToLastUpstreamRxByte:    e.getDurationFromProto(cp.GetTimeToLastUpstreamRxByte()),
		TimeToFirstDownstreamTxByte: e.getDurationFromProto(cp.GetTimeToFirstDownstreamTxByte()),
		TimeToLastDownstreamTxByte:  e.getDurationFromProto(cp.GetTimeToLastDownstreamTxByte()),
	}
}

func (e *HTTPAccessLogEntry) setSource() {
	srcAddress := e.entry.GetCommonProperties().GetDownstreamRemoteAddress().GetSocketAddress()
	e.Source = &RequestEndpoint{
		Address: &TCPAddr{
			IP:   srcAddress.GetAddress(),
			Port: int(srcAddress.GetPortValue()),
		},
		Metadata: make(map[string]string),
	}
}

func (e *HTTPAccessLogEntry) setSourceMetadata() error {
	sourceMetadata, err := GetMetadataAttributes(Headers(e.entry.GetRequest().GetRequestHeaders()).GetMetadata())
	if err != nil {
		return errors.WrapIf(err, "could not get source metadata")
	}
	e.Source.SetAttributes(sourceMetadata)

	return nil
}

func (e *HTTPAccessLogEntry) setDestination() {
	dstAddress := e.entry.GetCommonProperties().GetDownstreamLocalAddress().GetSocketAddress()
	e.Destination = &RequestEndpoint{
		Address: &TCPAddr{
			IP:   dstAddress.GetAddress(),
			Port: int(dstAddress.GetPortValue()),
		},
		Metadata: make(map[string]string),
	}

	e.Destination.Metadata["authority"] = e.entry.GetRequest().GetAuthority()
}

func (e *HTTPAccessLogEntry) setDestinationMetadata() error {
	destinationMetadata, err := GetMetadataAttributes(Headers(e.entry.GetResponse().GetResponseHeaders()).GetMetadata())
	if err != nil {
		return errors.WrapIf(err, "could not get destination metadata")
	}
	e.Destination.SetAttributes(destinationMetadata)

	return nil
}

func (e *HTTPAccessLogEntry) setAuthinfo() {
	values := make(map[string]interface{})
	info := &AuthInfo{}

	md := e.entry.GetCommonProperties().GetMetadata().GetFilterMetadata()
	if mdvalue, ok := md["istio_authn"]; ok {
		values = getStructValues(mdvalue)
	}

	if v, ok := values["request.auth.principal"]; ok {
		if value, ok := v.(string); ok {
			info.RequestPrincipal = value
		}
	}

	if v, ok := values["source.namespace"]; ok {
		if value, ok := v.(string); ok {
			info.Namespace = value
		}
	}

	if v, ok := values["source.principal"]; ok {
		if value, ok := v.(string); ok {
			info.Principal = value
		}
	}

	if v, ok := values["source.user"]; ok {
		if value, ok := v.(string); ok {
			info.User = value
		}
	}

	e.AuthInfo = info
}

func (e *HTTPAccessLogEntry) setRequest() {
	req := e.entry.GetRequest()
	e.Request = &HTTPRequest{
		ID:           req.GetRequestId(),
		Method:       req.GetRequestMethod().String(),
		Scheme:       req.GetScheme(),
		Authority:    req.GetAuthority(),
		Path:         req.GetPath(),
		UserAgent:    req.GetUserAgent(),
		Referer:      req.GetReferer(),
		ForwardedFor: req.GetForwardedFor(),
		OriginalPath: req.GetOriginalPath(),
		HeaderBytes:  req.GetRequestHeadersBytes(),
		BodyBytes:    req.GetRequestBodyBytes(),
		Headers:      Headers(req.GetRequestHeaders()).GetAllWithoutMetadataHeaders(),
	}
}

func (e *HTTPAccessLogEntry) setResponse() {
	resp := e.entry.GetResponse()
	e.Response = &HTTPResponse{
		StatusCode:        resp.GetResponseCode().GetValue(),
		StatusCodeDetails: resp.GetResponseCodeDetails(),
		HeaderBytes:       resp.GetResponseHeadersBytes(),
		BodyBytes:         resp.GetResponseBodyBytes(),
		Headers:           Headers(resp.GetResponseHeaders()).GetAllWithoutMetadataHeaders(),
		Trailers:          resp.GetResponseTrailers(),
		Flags:             e.getResponseFlags(),
	}
}

func (e *HTTPAccessLogEntry) getResponseFlags() []string {
	rf := e.entry.GetCommonProperties().GetResponseFlags()
	if rf == nil {
		return []string{"-"}
	}

	flags := make([]string, 0)
	checks := map[string]bool{
		"LH":   rf.FailedLocalHealthcheck,
		"UH":   rf.NoHealthyUpstream,
		"UT":   rf.UpstreamRequestTimeout,
		"LR":   rf.LocalReset,
		"UR":   rf.UpstreamRemoteReset,
		"UF":   rf.UpstreamConnectionFailure,
		"UC":   rf.UpstreamConnectionTermination,
		"UO":   rf.UpstreamOverflow,
		"NR":   rf.NoRouteFound,
		"DI":   rf.DelayInjected,
		"FI":   rf.FaultInjected,
		"RL":   rf.RateLimited,
		"RLSE": rf.RateLimitServiceError,
		"DC":   rf.DownstreamConnectionTermination,
		"URX":  rf.UpstreamRetryLimitExceeded,
		"SI":   rf.StreamIdleTimeout,
		"DPE":  rf.DownstreamProtocolError,
		"IH":   rf.InvalidEnvoyRequestHeaders,
	}

	for code, fn := range checks {
		if fn {
			flags = append(flags, code)
		}
	}

	if rf.UnauthorizedDetails.GetReason().String() != "" {
		flags = append(flags, "UAEX")
	}

	if len(flags) > 0 {
		return flags
	}

	return []string{"-"}
}
