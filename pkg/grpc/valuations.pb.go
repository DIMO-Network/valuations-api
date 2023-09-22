// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v4.24.3
// source: pkg/grpc/valuations.proto

package grpc

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type ValuationResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Total            float32 `protobuf:"fixed32,1,opt,name=total,proto3" json:"total,omitempty"`
	GrowthPercentage float32 `protobuf:"fixed32,2,opt,name=growthPercentage,proto3" json:"growthPercentage,omitempty"`
}

func (x *ValuationResponse) Reset() {
	*x = ValuationResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_grpc_valuations_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ValuationResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ValuationResponse) ProtoMessage() {}

func (x *ValuationResponse) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_grpc_valuations_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ValuationResponse.ProtoReflect.Descriptor instead.
func (*ValuationResponse) Descriptor() ([]byte, []int) {
	return file_pkg_grpc_valuations_proto_rawDescGZIP(), []int{0}
}

func (x *ValuationResponse) GetTotal() float32 {
	if x != nil {
		return x.Total
	}
	return 0
}

func (x *ValuationResponse) GetGrowthPercentage() float32 {
	if x != nil {
		return x.GrowthPercentage
	}
	return 0
}

type DeviceValuationRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	UserDeviceId string `protobuf:"bytes,1,opt,name=userDeviceId,proto3" json:"userDeviceId,omitempty"`
}

func (x *DeviceValuationRequest) Reset() {
	*x = DeviceValuationRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_grpc_valuations_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeviceValuationRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeviceValuationRequest) ProtoMessage() {}

func (x *DeviceValuationRequest) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_grpc_valuations_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeviceValuationRequest.ProtoReflect.Descriptor instead.
func (*DeviceValuationRequest) Descriptor() ([]byte, []int) {
	return file_pkg_grpc_valuations_proto_rawDescGZIP(), []int{1}
}

func (x *DeviceValuationRequest) GetUserDeviceId() string {
	if x != nil {
		return x.UserDeviceId
	}
	return ""
}

type DeviceOfferRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	UserDeviceId string `protobuf:"bytes,1,opt,name=userDeviceId,proto3" json:"userDeviceId,omitempty"`
}

func (x *DeviceOfferRequest) Reset() {
	*x = DeviceOfferRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_grpc_valuations_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeviceOfferRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeviceOfferRequest) ProtoMessage() {}

func (x *DeviceOfferRequest) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_grpc_valuations_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeviceOfferRequest.ProtoReflect.Descriptor instead.
func (*DeviceOfferRequest) Descriptor() ([]byte, []int) {
	return file_pkg_grpc_valuations_proto_rawDescGZIP(), []int{2}
}

func (x *DeviceOfferRequest) GetUserDeviceId() string {
	if x != nil {
		return x.UserDeviceId
	}
	return ""
}

type DeviceValuation struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ValuationSets []*ValuationSet `protobuf:"bytes,1,rep,name=valuationSets,proto3" json:"valuationSets,omitempty"`
}

func (x *DeviceValuation) Reset() {
	*x = DeviceValuation{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_grpc_valuations_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeviceValuation) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeviceValuation) ProtoMessage() {}

func (x *DeviceValuation) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_grpc_valuations_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeviceValuation.ProtoReflect.Descriptor instead.
func (*DeviceValuation) Descriptor() ([]byte, []int) {
	return file_pkg_grpc_valuations_proto_rawDescGZIP(), []int{3}
}

func (x *DeviceValuation) GetValuationSets() []*ValuationSet {
	if x != nil {
		return x.ValuationSets
	}
	return nil
}

type DeviceOffer struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	OfferSets []*OfferSet `protobuf:"bytes,1,rep,name=offerSets,proto3" json:"offerSets,omitempty"`
}

func (x *DeviceOffer) Reset() {
	*x = DeviceOffer{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_grpc_valuations_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeviceOffer) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeviceOffer) ProtoMessage() {}

func (x *DeviceOffer) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_grpc_valuations_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeviceOffer.ProtoReflect.Descriptor instead.
func (*DeviceOffer) Descriptor() ([]byte, []int) {
	return file_pkg_grpc_valuations_proto_rawDescGZIP(), []int{4}
}

func (x *DeviceOffer) GetOfferSets() []*OfferSet {
	if x != nil {
		return x.OfferSets
	}
	return nil
}

type ValuationSet struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Vendor           string `protobuf:"bytes,1,opt,name=vendor,proto3" json:"vendor,omitempty"`
	Updated          string `protobuf:"bytes,2,opt,name=updated,proto3" json:"updated,omitempty"`
	Mileage          int32  `protobuf:"varint,3,opt,name=mileage,proto3" json:"mileage,omitempty"`
	ZipCode          string `protobuf:"bytes,4,opt,name=zipCode,proto3" json:"zipCode,omitempty"`
	TradeInSource    string `protobuf:"bytes,5,opt,name=tradeInSource,proto3" json:"tradeInSource,omitempty"`
	TradeIn          int32  `protobuf:"varint,6,opt,name=tradeIn,proto3" json:"tradeIn,omitempty"`
	TradeInClean     int32  `protobuf:"varint,7,opt,name=tradeInClean,proto3" json:"tradeInClean,omitempty"`
	TradeInAverage   int32  `protobuf:"varint,8,opt,name=tradeInAverage,proto3" json:"tradeInAverage,omitempty"`
	TradeInRough     int32  `protobuf:"varint,9,opt,name=tradeInRough,proto3" json:"tradeInRough,omitempty"`
	RetailSource     string `protobuf:"bytes,10,opt,name=RetailSource,proto3" json:"RetailSource,omitempty"`
	Retail           int32  `protobuf:"varint,11,opt,name=retail,proto3" json:"retail,omitempty"`
	RetailClean      int32  `protobuf:"varint,12,opt,name=retailClean,proto3" json:"retailClean,omitempty"`
	RetailAverage    int32  `protobuf:"varint,13,opt,name=retailAverage,proto3" json:"retailAverage,omitempty"`
	RetailRough      int32  `protobuf:"varint,14,opt,name=retailRough,proto3" json:"retailRough,omitempty"`
	OdometerUnit     string `protobuf:"bytes,15,opt,name=odometerUnit,proto3" json:"odometerUnit,omitempty"`
	Odometer         int32  `protobuf:"varint,16,opt,name=odometer,proto3" json:"odometer,omitempty"`
	UserDisplayPrice int32  `protobuf:"varint,17,opt,name=userDisplayPrice,proto3" json:"userDisplayPrice,omitempty"`
	Currency         string `protobuf:"bytes,18,opt,name=currency,proto3" json:"currency,omitempty"`
}

func (x *ValuationSet) Reset() {
	*x = ValuationSet{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_grpc_valuations_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ValuationSet) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ValuationSet) ProtoMessage() {}

func (x *ValuationSet) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_grpc_valuations_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ValuationSet.ProtoReflect.Descriptor instead.
func (*ValuationSet) Descriptor() ([]byte, []int) {
	return file_pkg_grpc_valuations_proto_rawDescGZIP(), []int{5}
}

func (x *ValuationSet) GetVendor() string {
	if x != nil {
		return x.Vendor
	}
	return ""
}

func (x *ValuationSet) GetUpdated() string {
	if x != nil {
		return x.Updated
	}
	return ""
}

func (x *ValuationSet) GetMileage() int32 {
	if x != nil {
		return x.Mileage
	}
	return 0
}

func (x *ValuationSet) GetZipCode() string {
	if x != nil {
		return x.ZipCode
	}
	return ""
}

func (x *ValuationSet) GetTradeInSource() string {
	if x != nil {
		return x.TradeInSource
	}
	return ""
}

func (x *ValuationSet) GetTradeIn() int32 {
	if x != nil {
		return x.TradeIn
	}
	return 0
}

func (x *ValuationSet) GetTradeInClean() int32 {
	if x != nil {
		return x.TradeInClean
	}
	return 0
}

func (x *ValuationSet) GetTradeInAverage() int32 {
	if x != nil {
		return x.TradeInAverage
	}
	return 0
}

func (x *ValuationSet) GetTradeInRough() int32 {
	if x != nil {
		return x.TradeInRough
	}
	return 0
}

func (x *ValuationSet) GetRetailSource() string {
	if x != nil {
		return x.RetailSource
	}
	return ""
}

func (x *ValuationSet) GetRetail() int32 {
	if x != nil {
		return x.Retail
	}
	return 0
}

func (x *ValuationSet) GetRetailClean() int32 {
	if x != nil {
		return x.RetailClean
	}
	return 0
}

func (x *ValuationSet) GetRetailAverage() int32 {
	if x != nil {
		return x.RetailAverage
	}
	return 0
}

func (x *ValuationSet) GetRetailRough() int32 {
	if x != nil {
		return x.RetailRough
	}
	return 0
}

func (x *ValuationSet) GetOdometerUnit() string {
	if x != nil {
		return x.OdometerUnit
	}
	return ""
}

func (x *ValuationSet) GetOdometer() int32 {
	if x != nil {
		return x.Odometer
	}
	return 0
}

func (x *ValuationSet) GetUserDisplayPrice() int32 {
	if x != nil {
		return x.UserDisplayPrice
	}
	return 0
}

func (x *ValuationSet) GetCurrency() string {
	if x != nil {
		return x.Currency
	}
	return ""
}

type OfferSet struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Source  string   `protobuf:"bytes,1,opt,name=source,proto3" json:"source,omitempty"`
	Updated string   `protobuf:"bytes,2,opt,name=updated,proto3" json:"updated,omitempty"`
	Mileage int32    `protobuf:"varint,3,opt,name=mileage,proto3" json:"mileage,omitempty"`
	ZipCode string   `protobuf:"bytes,4,opt,name=zipCode,proto3" json:"zipCode,omitempty"`
	Offers  []*Offer `protobuf:"bytes,5,rep,name=offers,proto3" json:"offers,omitempty"`
}

func (x *OfferSet) Reset() {
	*x = OfferSet{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_grpc_valuations_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *OfferSet) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*OfferSet) ProtoMessage() {}

func (x *OfferSet) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_grpc_valuations_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use OfferSet.ProtoReflect.Descriptor instead.
func (*OfferSet) Descriptor() ([]byte, []int) {
	return file_pkg_grpc_valuations_proto_rawDescGZIP(), []int{6}
}

func (x *OfferSet) GetSource() string {
	if x != nil {
		return x.Source
	}
	return ""
}

func (x *OfferSet) GetUpdated() string {
	if x != nil {
		return x.Updated
	}
	return ""
}

func (x *OfferSet) GetMileage() int32 {
	if x != nil {
		return x.Mileage
	}
	return 0
}

func (x *OfferSet) GetZipCode() string {
	if x != nil {
		return x.ZipCode
	}
	return ""
}

func (x *OfferSet) GetOffers() []*Offer {
	if x != nil {
		return x.Offers
	}
	return nil
}

type Offer struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Vendor        string `protobuf:"bytes,1,opt,name=vendor,proto3" json:"vendor,omitempty"`
	Price         int32  `protobuf:"varint,2,opt,name=price,proto3" json:"price,omitempty"`
	Url           string `protobuf:"bytes,3,opt,name=url,proto3" json:"url,omitempty"`
	Error         string `protobuf:"bytes,4,opt,name=error,proto3" json:"error,omitempty"`
	Grade         string `protobuf:"bytes,5,opt,name=grade,proto3" json:"grade,omitempty"`
	DeclineReason string `protobuf:"bytes,6,opt,name=declineReason,proto3" json:"declineReason,omitempty"`
}

func (x *Offer) Reset() {
	*x = Offer{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_grpc_valuations_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Offer) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Offer) ProtoMessage() {}

func (x *Offer) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_grpc_valuations_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Offer.ProtoReflect.Descriptor instead.
func (*Offer) Descriptor() ([]byte, []int) {
	return file_pkg_grpc_valuations_proto_rawDescGZIP(), []int{7}
}

func (x *Offer) GetVendor() string {
	if x != nil {
		return x.Vendor
	}
	return ""
}

func (x *Offer) GetPrice() int32 {
	if x != nil {
		return x.Price
	}
	return 0
}

func (x *Offer) GetUrl() string {
	if x != nil {
		return x.Url
	}
	return ""
}

func (x *Offer) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

func (x *Offer) GetGrade() string {
	if x != nil {
		return x.Grade
	}
	return ""
}

func (x *Offer) GetDeclineReason() string {
	if x != nil {
		return x.DeclineReason
	}
	return ""
}

var File_pkg_grpc_valuations_proto protoreflect.FileDescriptor

var file_pkg_grpc_valuations_proto_rawDesc = []byte{
	0x0a, 0x19, 0x70, 0x6b, 0x67, 0x2f, 0x67, 0x72, 0x70, 0x63, 0x2f, 0x76, 0x61, 0x6c, 0x75, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0a, 0x76, 0x61, 0x6c,
	0x75, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x1a, 0x1b, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x65, 0x6d, 0x70, 0x74, 0x79, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x22, 0x55, 0x0a, 0x11, 0x56, 0x61, 0x6c, 0x75, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x74, 0x6f, 0x74,
	0x61, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x02, 0x52, 0x05, 0x74, 0x6f, 0x74, 0x61, 0x6c, 0x12,
	0x2a, 0x0a, 0x10, 0x67, 0x72, 0x6f, 0x77, 0x74, 0x68, 0x50, 0x65, 0x72, 0x63, 0x65, 0x6e, 0x74,
	0x61, 0x67, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x02, 0x52, 0x10, 0x67, 0x72, 0x6f, 0x77, 0x74,
	0x68, 0x50, 0x65, 0x72, 0x63, 0x65, 0x6e, 0x74, 0x61, 0x67, 0x65, 0x22, 0x3c, 0x0a, 0x16, 0x44,
	0x65, 0x76, 0x69, 0x63, 0x65, 0x56, 0x61, 0x6c, 0x75, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x22, 0x0a, 0x0c, 0x75, 0x73, 0x65, 0x72, 0x44, 0x65, 0x76,
	0x69, 0x63, 0x65, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x75, 0x73, 0x65,
	0x72, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x49, 0x64, 0x22, 0x38, 0x0a, 0x12, 0x44, 0x65, 0x76,
	0x69, 0x63, 0x65, 0x4f, 0x66, 0x66, 0x65, 0x72, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12,
	0x22, 0x0a, 0x0c, 0x75, 0x73, 0x65, 0x72, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x49, 0x64, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x75, 0x73, 0x65, 0x72, 0x44, 0x65, 0x76, 0x69, 0x63,
	0x65, 0x49, 0x64, 0x22, 0x51, 0x0a, 0x0f, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x56, 0x61, 0x6c,
	0x75, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x3e, 0x0a, 0x0d, 0x76, 0x61, 0x6c, 0x75, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x53, 0x65, 0x74, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x18, 0x2e,
	0x76, 0x61, 0x6c, 0x75, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x56, 0x61, 0x6c, 0x75, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x53, 0x65, 0x74, 0x52, 0x0d, 0x76, 0x61, 0x6c, 0x75, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x53, 0x65, 0x74, 0x73, 0x22, 0x41, 0x0a, 0x0b, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65,
	0x4f, 0x66, 0x66, 0x65, 0x72, 0x12, 0x32, 0x0a, 0x09, 0x6f, 0x66, 0x66, 0x65, 0x72, 0x53, 0x65,
	0x74, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x76, 0x61, 0x6c, 0x75, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x4f, 0x66, 0x66, 0x65, 0x72, 0x53, 0x65, 0x74, 0x52, 0x09,
	0x6f, 0x66, 0x66, 0x65, 0x72, 0x53, 0x65, 0x74, 0x73, 0x22, 0xd2, 0x04, 0x0a, 0x0c, 0x56, 0x61,
	0x6c, 0x75, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x53, 0x65, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x76, 0x65,
	0x6e, 0x64, 0x6f, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x76, 0x65, 0x6e, 0x64,
	0x6f, 0x72, 0x12, 0x18, 0x0a, 0x07, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x07, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x12, 0x18, 0x0a, 0x07,
	0x6d, 0x69, 0x6c, 0x65, 0x61, 0x67, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x07, 0x6d,
	0x69, 0x6c, 0x65, 0x61, 0x67, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x7a, 0x69, 0x70, 0x43, 0x6f, 0x64,
	0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x7a, 0x69, 0x70, 0x43, 0x6f, 0x64, 0x65,
	0x12, 0x24, 0x0a, 0x0d, 0x74, 0x72, 0x61, 0x64, 0x65, 0x49, 0x6e, 0x53, 0x6f, 0x75, 0x72, 0x63,
	0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x74, 0x72, 0x61, 0x64, 0x65, 0x49, 0x6e,
	0x53, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x74, 0x72, 0x61, 0x64, 0x65, 0x49,
	0x6e, 0x18, 0x06, 0x20, 0x01, 0x28, 0x05, 0x52, 0x07, 0x74, 0x72, 0x61, 0x64, 0x65, 0x49, 0x6e,
	0x12, 0x22, 0x0a, 0x0c, 0x74, 0x72, 0x61, 0x64, 0x65, 0x49, 0x6e, 0x43, 0x6c, 0x65, 0x61, 0x6e,
	0x18, 0x07, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0c, 0x74, 0x72, 0x61, 0x64, 0x65, 0x49, 0x6e, 0x43,
	0x6c, 0x65, 0x61, 0x6e, 0x12, 0x26, 0x0a, 0x0e, 0x74, 0x72, 0x61, 0x64, 0x65, 0x49, 0x6e, 0x41,
	0x76, 0x65, 0x72, 0x61, 0x67, 0x65, 0x18, 0x08, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0e, 0x74, 0x72,
	0x61, 0x64, 0x65, 0x49, 0x6e, 0x41, 0x76, 0x65, 0x72, 0x61, 0x67, 0x65, 0x12, 0x22, 0x0a, 0x0c,
	0x74, 0x72, 0x61, 0x64, 0x65, 0x49, 0x6e, 0x52, 0x6f, 0x75, 0x67, 0x68, 0x18, 0x09, 0x20, 0x01,
	0x28, 0x05, 0x52, 0x0c, 0x74, 0x72, 0x61, 0x64, 0x65, 0x49, 0x6e, 0x52, 0x6f, 0x75, 0x67, 0x68,
	0x12, 0x22, 0x0a, 0x0c, 0x52, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x53, 0x6f, 0x75, 0x72, 0x63, 0x65,
	0x18, 0x0a, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x52, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x53, 0x6f,
	0x75, 0x72, 0x63, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x72, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x18, 0x0b,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x06, 0x72, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x12, 0x20, 0x0a, 0x0b,
	0x72, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x43, 0x6c, 0x65, 0x61, 0x6e, 0x18, 0x0c, 0x20, 0x01, 0x28,
	0x05, 0x52, 0x0b, 0x72, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x43, 0x6c, 0x65, 0x61, 0x6e, 0x12, 0x24,
	0x0a, 0x0d, 0x72, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x41, 0x76, 0x65, 0x72, 0x61, 0x67, 0x65, 0x18,
	0x0d, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0d, 0x72, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x41, 0x76, 0x65,
	0x72, 0x61, 0x67, 0x65, 0x12, 0x20, 0x0a, 0x0b, 0x72, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x52, 0x6f,
	0x75, 0x67, 0x68, 0x18, 0x0e, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0b, 0x72, 0x65, 0x74, 0x61, 0x69,
	0x6c, 0x52, 0x6f, 0x75, 0x67, 0x68, 0x12, 0x22, 0x0a, 0x0c, 0x6f, 0x64, 0x6f, 0x6d, 0x65, 0x74,
	0x65, 0x72, 0x55, 0x6e, 0x69, 0x74, 0x18, 0x0f, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x6f, 0x64,
	0x6f, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x55, 0x6e, 0x69, 0x74, 0x12, 0x1a, 0x0a, 0x08, 0x6f, 0x64,
	0x6f, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x18, 0x10, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x6f, 0x64,
	0x6f, 0x6d, 0x65, 0x74, 0x65, 0x72, 0x12, 0x2a, 0x0a, 0x10, 0x75, 0x73, 0x65, 0x72, 0x44, 0x69,
	0x73, 0x70, 0x6c, 0x61, 0x79, 0x50, 0x72, 0x69, 0x63, 0x65, 0x18, 0x11, 0x20, 0x01, 0x28, 0x05,
	0x52, 0x10, 0x75, 0x73, 0x65, 0x72, 0x44, 0x69, 0x73, 0x70, 0x6c, 0x61, 0x79, 0x50, 0x72, 0x69,
	0x63, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x63, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x63, 0x79, 0x18, 0x12,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x63, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x63, 0x79, 0x22, 0x9b,
	0x01, 0x0a, 0x08, 0x4f, 0x66, 0x66, 0x65, 0x72, 0x53, 0x65, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x73,
	0x6f, 0x75, 0x72, 0x63, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x6f, 0x75,
	0x72, 0x63, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x12, 0x18, 0x0a,
	0x07, 0x6d, 0x69, 0x6c, 0x65, 0x61, 0x67, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52, 0x07,
	0x6d, 0x69, 0x6c, 0x65, 0x61, 0x67, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x7a, 0x69, 0x70, 0x43, 0x6f,
	0x64, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x7a, 0x69, 0x70, 0x43, 0x6f, 0x64,
	0x65, 0x12, 0x29, 0x0a, 0x06, 0x6f, 0x66, 0x66, 0x65, 0x72, 0x73, 0x18, 0x05, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x11, 0x2e, 0x76, 0x61, 0x6c, 0x75, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x4f,
	0x66, 0x66, 0x65, 0x72, 0x52, 0x06, 0x6f, 0x66, 0x66, 0x65, 0x72, 0x73, 0x22, 0x99, 0x01, 0x0a,
	0x05, 0x4f, 0x66, 0x66, 0x65, 0x72, 0x12, 0x16, 0x0a, 0x06, 0x76, 0x65, 0x6e, 0x64, 0x6f, 0x72,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x76, 0x65, 0x6e, 0x64, 0x6f, 0x72, 0x12, 0x14,
	0x0a, 0x05, 0x70, 0x72, 0x69, 0x63, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x05, 0x70,
	0x72, 0x69, 0x63, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x75, 0x72, 0x6c, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x03, 0x75, 0x72, 0x6c, 0x12, 0x14, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x18,
	0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x12, 0x14, 0x0a, 0x05,
	0x67, 0x72, 0x61, 0x64, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x67, 0x72, 0x61,
	0x64, 0x65, 0x12, 0x24, 0x0a, 0x0d, 0x64, 0x65, 0x63, 0x6c, 0x69, 0x6e, 0x65, 0x52, 0x65, 0x61,
	0x73, 0x6f, 0x6e, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x64, 0x65, 0x63, 0x6c, 0x69,
	0x6e, 0x65, 0x52, 0x65, 0x61, 0x73, 0x6f, 0x6e, 0x32, 0xdc, 0x02, 0x0a, 0x11, 0x56, 0x61, 0x6c,
	0x75, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x49,
	0x0a, 0x10, 0x47, 0x65, 0x74, 0x41, 0x6c, 0x6c, 0x56, 0x61, 0x6c, 0x75, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x73, 0x12, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x1d, 0x2e, 0x76, 0x61, 0x6c,
	0x75, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x56, 0x61, 0x6c, 0x75, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x59, 0x0a, 0x16, 0x47, 0x65, 0x74,
	0x55, 0x73, 0x65, 0x72, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x56, 0x61, 0x6c, 0x75, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x12, 0x22, 0x2e, 0x76, 0x61, 0x6c, 0x75, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73,
	0x2e, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x56, 0x61, 0x6c, 0x75, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1b, 0x2e, 0x76, 0x61, 0x6c, 0x75, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x56, 0x61, 0x6c, 0x75, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x12, 0x4d, 0x0a, 0x12, 0x47, 0x65, 0x74, 0x55, 0x73, 0x65, 0x72, 0x44,
	0x65, 0x76, 0x69, 0x63, 0x65, 0x4f, 0x66, 0x66, 0x65, 0x72, 0x12, 0x1e, 0x2e, 0x76, 0x61, 0x6c,
	0x75, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x4f, 0x66,
	0x66, 0x65, 0x72, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x17, 0x2e, 0x76, 0x61, 0x6c,
	0x75, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x4f, 0x66,
	0x66, 0x65, 0x72, 0x12, 0x52, 0x0a, 0x19, 0x47, 0x65, 0x74, 0x41, 0x6c, 0x6c, 0x55, 0x73, 0x65,
	0x72, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x56, 0x61, 0x6c, 0x75, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x12, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x1d, 0x2e, 0x76, 0x61, 0x6c, 0x75, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x56, 0x61, 0x6c, 0x75, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x31, 0x5a, 0x2f, 0x67, 0x69, 0x74, 0x68, 0x75,
	0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x44, 0x49, 0x4d, 0x4f, 0x2d, 0x4e, 0x65, 0x74, 0x77, 0x6f,
	0x72, 0x6b, 0x2f, 0x76, 0x61, 0x6c, 0x75, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2d, 0x61, 0x70,
	0x69, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x67, 0x72, 0x70, 0x63, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_pkg_grpc_valuations_proto_rawDescOnce sync.Once
	file_pkg_grpc_valuations_proto_rawDescData = file_pkg_grpc_valuations_proto_rawDesc
)

func file_pkg_grpc_valuations_proto_rawDescGZIP() []byte {
	file_pkg_grpc_valuations_proto_rawDescOnce.Do(func() {
		file_pkg_grpc_valuations_proto_rawDescData = protoimpl.X.CompressGZIP(file_pkg_grpc_valuations_proto_rawDescData)
	})
	return file_pkg_grpc_valuations_proto_rawDescData
}

var file_pkg_grpc_valuations_proto_msgTypes = make([]protoimpl.MessageInfo, 8)
var file_pkg_grpc_valuations_proto_goTypes = []interface{}{
	(*ValuationResponse)(nil),      // 0: valuations.ValuationResponse
	(*DeviceValuationRequest)(nil), // 1: valuations.DeviceValuationRequest
	(*DeviceOfferRequest)(nil),     // 2: valuations.DeviceOfferRequest
	(*DeviceValuation)(nil),        // 3: valuations.DeviceValuation
	(*DeviceOffer)(nil),            // 4: valuations.DeviceOffer
	(*ValuationSet)(nil),           // 5: valuations.ValuationSet
	(*OfferSet)(nil),               // 6: valuations.OfferSet
	(*Offer)(nil),                  // 7: valuations.Offer
	(*emptypb.Empty)(nil),          // 8: google.protobuf.Empty
}
var file_pkg_grpc_valuations_proto_depIdxs = []int32{
	5, // 0: valuations.DeviceValuation.valuationSets:type_name -> valuations.ValuationSet
	6, // 1: valuations.DeviceOffer.offerSets:type_name -> valuations.OfferSet
	7, // 2: valuations.OfferSet.offers:type_name -> valuations.Offer
	8, // 3: valuations.ValuationsService.GetAllValuations:input_type -> google.protobuf.Empty
	1, // 4: valuations.ValuationsService.GetUserDeviceValuation:input_type -> valuations.DeviceValuationRequest
	2, // 5: valuations.ValuationsService.GetUserDeviceOffer:input_type -> valuations.DeviceOfferRequest
	8, // 6: valuations.ValuationsService.GetAllUserDeviceValuation:input_type -> google.protobuf.Empty
	0, // 7: valuations.ValuationsService.GetAllValuations:output_type -> valuations.ValuationResponse
	3, // 8: valuations.ValuationsService.GetUserDeviceValuation:output_type -> valuations.DeviceValuation
	4, // 9: valuations.ValuationsService.GetUserDeviceOffer:output_type -> valuations.DeviceOffer
	0, // 10: valuations.ValuationsService.GetAllUserDeviceValuation:output_type -> valuations.ValuationResponse
	7, // [7:11] is the sub-list for method output_type
	3, // [3:7] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_pkg_grpc_valuations_proto_init() }
func file_pkg_grpc_valuations_proto_init() {
	if File_pkg_grpc_valuations_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_pkg_grpc_valuations_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ValuationResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_pkg_grpc_valuations_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeviceValuationRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_pkg_grpc_valuations_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeviceOfferRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_pkg_grpc_valuations_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeviceValuation); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_pkg_grpc_valuations_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeviceOffer); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_pkg_grpc_valuations_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ValuationSet); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_pkg_grpc_valuations_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*OfferSet); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_pkg_grpc_valuations_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Offer); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_pkg_grpc_valuations_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   8,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_pkg_grpc_valuations_proto_goTypes,
		DependencyIndexes: file_pkg_grpc_valuations_proto_depIdxs,
		MessageInfos:      file_pkg_grpc_valuations_proto_msgTypes,
	}.Build()
	File_pkg_grpc_valuations_proto = out.File
	file_pkg_grpc_valuations_proto_rawDesc = nil
	file_pkg_grpc_valuations_proto_goTypes = nil
	file_pkg_grpc_valuations_proto_depIdxs = nil
}
