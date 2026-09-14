package main

import (
	"bytes"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	stdNetHTTPDefaultTimeoutMillis int64 = 30000
	stdNetHTTPDefaultMaxBytes      int64 = 1000000
	stdNetHTTPMaxDurationMillis    int64 = (1<<63 - 1) / int64(time.Millisecond)
	stdNetHTTPMaxResponseBytes     int64 = 1 << 62
)

type stdNetHTTPRequestOptions struct {
	Method          string
	Headers         [][2]string
	Body            []byte
	HasBody         bool
	UseBytes        bool
	TimeoutMillis   int64
	MaxBytes        int64
	FollowRedirects bool
}

func NewStdNetHTTPMap(runtime *RuntimeContext) Value {
	if runtime == nil {
		runtime = NewRuntimeContext(nil)
	}

	entries := map[string]Binding{
		"REQUEST":                   NewImmutableBinding(NewBuiltinFunctionValue(builtinStdNetHTTPRequest)),
		"REQUEST_OPTIONS_CONTRACT":  NewImmutableBinding(NewContractValue(stdNetHTTPRequestOptionsContract())),
		"REQUEST_RESPONSE_CONTRACT": NewImmutableBinding(NewContractValue(stdNetHTTPRequestResponseContract())),
	}

	return NewMapValue(entries, true)
}

func builtinStdNetHTTPRequest(_ *RuntimeContext, args []Value) (Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return Value{}, fmt.Errorf("NET.HTTP.REQUEST expected 1 or 2 arguments, got %d", len(args))
	}

	url, err := stdNetHTTPStringArg("NET.HTTP.REQUEST", args[0], 1)
	if err != nil {
		return Value{}, err
	}
	if strings.TrimSpace(url) == "" {
		return Value{}, fmt.Errorf("NET.HTTP.REQUEST argument 1 expected a non-empty URL")
	}

	options, err := stdNetHTTPRequestOptionsArg(args)
	if err != nil {
		return Value{}, err
	}

	var body io.Reader
	if options.HasBody {
		body = bytes.NewReader(options.Body)
	}

	request, err := http.NewRequest(options.Method, url, body)
	if err != nil {
		return Value{}, err
	}

	for _, header := range options.Headers {
		request.Header.Add(header[0], header[1])
	}

	client := &http.Client{
		Timeout: time.Duration(options.TimeoutMillis) * time.Millisecond,
	}
	if !options.FollowRedirects {
		client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	response, err := client.Do(request)
	if err != nil {
		return Value{}, err
	}
	defer response.Body.Close()

	data, err := stdNetHTTPReadResponseBody(response.Body, options.MaxBytes)
	if err != nil {
		return Value{}, err
	}

	var bodyValue Value
	if options.UseBytes {
		bodyValue = stdFileBundledBytesValue(data)
	} else {
		if !utf8.Valid(data) {
			return Value{}, fmt.Errorf("NET.HTTP.REQUEST text mode expected valid UTF-8 response body")
		}
		bodyValue = NewStringValue(string(data))
	}

	result := stdNetHTTPResponseValue(response, bodyValue)
	if err := stdNetHTTPRequestResponseContract().Check("NET.HTTP.REQUEST response", result); err != nil {
		return Value{}, err
	}

	return result, nil
}

func stdNetHTTPRequestOptionsArg(args []Value) (stdNetHTTPRequestOptions, error) {
	options := stdNetHTTPRequestOptions{
		Method:          "GET",
		TimeoutMillis:   stdNetHTTPDefaultTimeoutMillis,
		MaxBytes:        stdNetHTTPDefaultMaxBytes,
		FollowRedirects: true,
	}

	if len(args) < 2 {
		return options, nil
	}

	value := resolveSpecializedValue(args[1])
	if value.Kind == ValueVoid {
		return options, nil
	}
	if value.Kind != ValueMap || value.Map == nil {
		return options, fmt.Errorf("NET.HTTP.REQUEST argument 2 expected an options map or _")
	}

	if err := stdNetHTTPRequestOptionsContract().Check("NET.HTTP.REQUEST argument 2", value); err != nil {
		return options, err
	}

	if methodValue, ok := value.Map.Get("METHOD"); ok {
		method, err := stdNetHTTPStringOption("METHOD", methodValue.Value)
		if err != nil {
			return options, err
		}

		method = strings.TrimSpace(method)
		if method == "" {
			return options, fmt.Errorf("NET.HTTP.REQUEST option METHOD expected a non-empty string")
		}
		options.Method = method
	}

	if headersValue, ok := value.Map.Get("HEADERS"); ok {
		headers, err := stdNetHTTPHeadersOption(headersValue.Value)
		if err != nil {
			return options, err
		}
		options.Headers = headers
	}

	if bodyValue, ok := value.Map.Get("BODY"); ok {
		body, hasBody, err := stdNetHTTPBodyOption(bodyValue.Value)
		if err != nil {
			return options, err
		}
		options.Body = body
		options.HasBody = hasBody
	}

	if useBytesValue, ok := value.Map.Get("USE_BYTES"); ok {
		useBytes, err := stdNetHTTPBoolOption("USE_BYTES", useBytesValue.Value)
		if err != nil {
			return options, err
		}
		options.UseBytes = useBytes
	}

	if timeoutValue, ok := value.Map.Get("TIMEOUT_MILLI"); ok {
		timeoutMillis, err := stdNetHTTPInt64Option("TIMEOUT_MILLI", timeoutValue.Value, stdNetHTTPMaxDurationMillis)
		if err != nil {
			return options, err
		}
		options.TimeoutMillis = timeoutMillis
	}

	if maxBytesValue, ok := value.Map.Get("MAX_BYTES"); ok {
		maxBytes, err := stdNetHTTPInt64Option("MAX_BYTES", maxBytesValue.Value, stdNetHTTPMaxResponseBytes)
		if err != nil {
			return options, err
		}
		options.MaxBytes = maxBytes
	}

	if followRedirectsValue, ok := value.Map.Get("FOLLOW_REDIRECTS"); ok {
		followRedirects, err := stdNetHTTPBoolOption("FOLLOW_REDIRECTS", followRedirectsValue.Value)
		if err != nil {
			return options, err
		}
		options.FollowRedirects = followRedirects
	}

	return options, nil
}

func stdNetHTTPRequestOptionsContract() BindingContract {
	stringContract := stdNetHTTPExactContract(ValueString)
	numberContract := stdNetHTTPExactContract(ValueNumber)
	boolContract := stdNetHTTPExactContract(ValueBool)
	voidContract := stdNetHTTPExactContract(ValueVoid)
	headerPairContract := BindingContract{
		Kind:    BindingContractArrayKind,
		Element: stringContract.ClonePointer(),
	}
	headersContract := BindingContract{
		Kind:    BindingContractArrayKind,
		Element: headerPairContract.ClonePointer(),
	}
	bodyBytesContract := BindingContract{
		Kind:    BindingContractArrayKind,
		Element: numberContract.ClonePointer(),
	}
	bodyContract := stdNetHTTPUnionContract(
		stringContract,
		bodyBytesContract,
		voidContract,
	)

	return BindingContract{
		Kind:            BindingContractMapKind,
		IsStructuredMap: true,
		MapFields: []BindingContractMapField{
			{Key: "BODY", Contract: bodyContract, IsOptional: true},
			{Key: "FOLLOW_REDIRECTS", Contract: boolContract, IsOptional: true},
			{Key: "HEADERS", Contract: headersContract, IsOptional: true},
			{Key: "MAX_BYTES", Contract: numberContract, IsOptional: true},
			{Key: "METHOD", Contract: stringContract, IsOptional: true},
			{Key: "TIMEOUT_MILLI", Contract: numberContract, IsOptional: true},
			{Key: "USE_BYTES", Contract: boolContract, IsOptional: true},
		},
	}
}

func stdNetHTTPRequestResponseContract() BindingContract {
	stringContract := stdNetHTTPExactContract(ValueString)
	numberContract := stdNetHTTPExactContract(ValueNumber)
	headerPairContract := BindingContract{
		Kind:    BindingContractArrayKind,
		Element: stringContract.ClonePointer(),
	}
	headersContract := BindingContract{
		Kind:    BindingContractArrayKind,
		Element: headerPairContract.ClonePointer(),
	}
	bodyBytesContract := BindingContract{
		Kind:    BindingContractArrayKind,
		Element: numberContract.ClonePointer(),
	}
	bodyContract := stdNetHTTPUnionContract(
		stringContract,
		bodyBytesContract,
	)

	return BindingContract{
		Kind:            BindingContractMapKind,
		IsStructuredMap: true,
		MapFields: []BindingContractMapField{
			{Key: "BODY", Contract: bodyContract},
			{Key: "HEADERS", Contract: headersContract},
			{Key: "STATUS", Contract: numberContract},
			{Key: "URL", Contract: stringContract},
		},
	}
}

func stdNetHTTPExactContract(kind ValueKind) BindingContract {
	return BindingContract{
		Kind:      BindingContractExactKind,
		KindValue: kind,
	}
}

func stdNetHTTPUnionContract(options ...BindingContract) BindingContract {
	cloned := make([]BindingContract, len(options))
	for index := range options {
		cloned[index] = options[index].Clone()
	}

	return BindingContract{
		Kind:    BindingContractUnionKind,
		Options: cloned,
	}
}

func stdNetHTTPStringArg(name string, arg Value, position int) (string, error) {
	value := resolveSpecializedValue(arg)
	if value.Kind != ValueString || value.Text == nil {
		return "", fmt.Errorf("%s argument %d expected a string", name, position)
	}

	return value.Text.String(), nil
}

func stdNetHTTPStringOption(option string, arg Value) (string, error) {
	value := resolveSpecializedValue(arg)
	if value.Kind != ValueString || value.Text == nil {
		return "", fmt.Errorf("NET.HTTP.REQUEST option %s expected a string", option)
	}

	return value.Text.String(), nil
}

func stdNetHTTPBoolOption(option string, arg Value) (bool, error) {
	value := resolveSpecializedValue(arg)
	if value.Kind == ValueVoid {
		return false, nil
	}
	if value.Kind != ValueBool {
		return false, fmt.Errorf("NET.HTTP.REQUEST option %s expected a boolean or _", option)
	}

	return value.Bool, nil
}

func stdNetHTTPInt64Option(option string, arg Value, maximum int64) (int64, error) {
	value := resolveSpecializedValue(arg)
	if value.Kind != ValueNumber || value.Number == nil {
		return 0, fmt.Errorf("NET.HTTP.REQUEST option %s expected a non-negative whole number", option)
	}

	integer, accuracy := value.Number.Int64()
	if accuracy != big.Exact || integer < 0 {
		return 0, fmt.Errorf("NET.HTTP.REQUEST option %s expected a non-negative whole number", option)
	}
	if integer > maximum {
		return 0, fmt.Errorf("NET.HTTP.REQUEST option %s is too large", option)
	}

	return integer, nil
}

func stdNetHTTPHeadersOption(arg Value) ([][2]string, error) {
	value := resolveSpecializedValue(arg)
	if value.Kind == ValueVoid {
		return nil, nil
	}
	if value.Kind != ValueArray || value.Array == nil {
		return nil, fmt.Errorf("NET.HTTP.REQUEST option HEADERS expected an array of two-string arrays")
	}

	headers := make([][2]string, 0, value.Array.capacityHintHostLimited(int(arrayOrdinaryLimit)))
	if err := value.Array.ForEach(func(index *Number, item Value) error {
		header, err := stdNetHTTPHeaderPair(item)
		if err != nil {
			return fmt.Errorf("NET.HTTP.REQUEST option HEADERS item at index %s: %w", index.String(), err)
		}
		headers = append(headers, header)
		return nil
	}); err != nil {
		return nil, err
	}

	return headers, nil
}

func stdNetHTTPHeaderPair(arg Value) ([2]string, error) {
	value := resolveSpecializedValue(arg)
	if value.Kind != ValueArray || value.Array == nil {
		return [2]string{}, fmt.Errorf("expected a two-string array")
	}

	length, accuracy := value.Array.Len().Int64()
	if accuracy != big.Exact || length != 2 {
		return [2]string{}, fmt.Errorf("expected a two-string array")
	}

	keyValue, ok, err := value.Array.Get(NewSmallNumber(0))
	if err != nil {
		return [2]string{}, err
	}
	if !ok {
		return [2]string{}, fmt.Errorf("expected a two-string array")
	}

	valueValue, ok, err := value.Array.Get(NewSmallNumber(1))
	if err != nil {
		return [2]string{}, err
	}
	if !ok {
		return [2]string{}, fmt.Errorf("expected a two-string array")
	}

	key, err := stdNetHTTPHeaderString(keyValue)
	if err != nil {
		return [2]string{}, err
	}
	if key == "" {
		return [2]string{}, fmt.Errorf("expected a non-empty header key")
	}

	headerValue, err := stdNetHTTPHeaderString(valueValue)
	if err != nil {
		return [2]string{}, err
	}

	return [2]string{key, headerValue}, nil
}

func stdNetHTTPHeaderString(arg Value) (string, error) {
	value := resolveSpecializedValue(arg)
	if value.Kind != ValueString || value.Text == nil {
		return "", fmt.Errorf("expected a two-string array")
	}

	text := value.Text.String()
	if strings.ContainsAny(text, "\r\n") {
		return "", fmt.Errorf("header strings must not contain newlines")
	}

	return text, nil
}

func stdNetHTTPBodyOption(arg Value) ([]byte, bool, error) {
	value := resolveSpecializedValue(arg)
	if value.Kind == ValueVoid {
		return nil, false, nil
	}

	switch value.Kind {
	case ValueString:
		if value.Text == nil {
			return nil, false, fmt.Errorf("NET.HTTP.REQUEST option BODY expected a string, byte array or _")
		}

		return []byte(value.Text.String()), true, nil

	case ValueArray:
		bytes, err := stdNetHTTPByteArrayOption(value)
		if err != nil {
			return nil, false, err
		}

		return bytes, true, nil

	default:
		return nil, false, fmt.Errorf("NET.HTTP.REQUEST option BODY expected a string, byte array or _")
	}
}

func stdNetHTTPByteArrayOption(value Value) ([]byte, error) {
	if value.Array == nil {
		return nil, fmt.Errorf("NET.HTTP.REQUEST option BODY expected a valid byte array")
	}

	bytes := make([]byte, 0)
	if err := value.Array.ForEach(func(index *Number, item Value) error {
		item = resolveSpecializedValue(item)
		if item.Kind != ValueNumber || item.Number == nil {
			return fmt.Errorf("NET.HTTP.REQUEST option BODY byte array item at index %s expected a number from 0 to 255", index.String())
		}

		integer, accuracy := item.Number.Int64()
		if accuracy != big.Exact || integer < 0 || integer > 255 {
			return fmt.Errorf("NET.HTTP.REQUEST option BODY byte array item at index %s expected a whole number from 0 to 255", index.String())
		}

		bytes = append(bytes, byte(integer))
		return nil
	}); err != nil {
		return nil, err
	}

	return bytes, nil
}

func stdNetHTTPReadResponseBody(reader io.Reader, maxBytes int64) ([]byte, error) {
	limited := io.LimitReader(reader, maxBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maxBytes {
		return nil, fmt.Errorf("NET.HTTP.REQUEST response body exceeded MAX_BYTES")
	}

	return data, nil
}

func stdNetHTTPResponseValue(response *http.Response, body Value) Value {
	return NewMapValue(map[string]Binding{
		"BODY":    NewImmutableBinding(body),
		"HEADERS": NewImmutableBinding(stdNetHTTPHeaderArrayValue(response.Header)),
		"STATUS":  NewImmutableBinding(NewNumberValueFromInt(response.StatusCode)),
		"URL":     NewImmutableBinding(NewStringValue(response.Request.URL.String())),
	}, false)
}

func stdNetHTTPHeaderArrayValue(headers http.Header) Value {
	keys := make([]string, 0, len(headers))
	for key := range headers {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	pairs := make([]Value, 0)
	for _, key := range keys {
		for _, value := range headers.Values(key) {
			pairs = append(pairs, NewArrayValue([]Value{
				NewStringValue(key),
				NewStringValue(value),
			}, false))
		}
	}

	return NewArrayValue(pairs, false)
}
