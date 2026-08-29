package api

import (
	"errors"
	"net"
	"net/http"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

const (
	objectListPermissionMessage     = "Provider 无权访问此 OSS 桶，请检查生产权限配置"
	objectListMissingBucketMessage  = "OSS 桶不存在或当前 Provider 无权查看"
	objectListInvalidRequestMessage = "文件列表分页参数无效，请刷新后重试"
	objectListEndpointMessage       = "OSS 区域或访问端点配置不匹配"
	objectListUnavailableMessage    = "OSS 暂时不可用，请稍后重试"
	objectListUnknownMessage        = "文件列表加载失败，请联系管理员"
)

func classifyObjectListError(err error) (int, string) {
	if err == nil {
		return 0, ""
	}

	if serviceErr, ok := asOSSServiceError(err); ok {
		switch serviceErr.Code {
		case "AccessDenied", "InvalidAccessKeyId", "SignatureDoesNotMatch",
			"InvalidSecurityToken", "SecurityTokenExpired", "MissingSecurityToken":
			return http.StatusServiceUnavailable, objectListPermissionMessage
		case "NoSuchBucket":
			return http.StatusNotFound, objectListMissingBucketMessage
		case "InvalidArgument", "InvalidBucketName", "InvalidMarker", "InvalidMaxKeys":
			return http.StatusBadRequest, objectListInvalidRequestMessage
		case "PermanentRedirect", "IncorrectEndpoint", "AuthorizationHeaderMalformed", "InvalidRegion":
			return http.StatusServiceUnavailable, objectListEndpointMessage
		}

		switch {
		case serviceErr.StatusCode == http.StatusNotFound:
			return http.StatusNotFound, objectListMissingBucketMessage
		case serviceErr.StatusCode == http.StatusBadRequest:
			return http.StatusBadRequest, objectListInvalidRequestMessage
		case serviceErr.StatusCode == http.StatusForbidden:
			return http.StatusServiceUnavailable, objectListPermissionMessage
		case serviceErr.StatusCode >= http.StatusInternalServerError:
			return http.StatusBadGateway, objectListUnavailableMessage
		}
	}

	var unexpectedStatus oss.UnexpectedStatusCodeError
	if errors.As(err, &unexpectedStatus) && unexpectedStatus.Got() >= http.StatusMultipleChoices && unexpectedStatus.Got() < http.StatusBadRequest {
		return http.StatusServiceUnavailable, objectListEndpointMessage
	}

	var networkErr net.Error
	if errors.As(err, &networkErr) {
		return http.StatusBadGateway, objectListUnavailableMessage
	}

	return http.StatusInternalServerError, objectListUnknownMessage
}

func asOSSServiceError(err error) (oss.ServiceError, bool) {
	var serviceErr oss.ServiceError
	if errors.As(err, &serviceErr) {
		return serviceErr, true
	}

	var serviceErrPtr *oss.ServiceError
	if errors.As(err, &serviceErrPtr) && serviceErrPtr != nil {
		return *serviceErrPtr, true
	}

	return oss.ServiceError{}, false
}
