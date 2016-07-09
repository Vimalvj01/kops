package awstasks

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/kops/upup/pkg/fi"
)

type BlockDeviceMapping struct {
	VirtualName *string

	EbsDeleteOnTermination *bool
	EbsVolumeSize          *int64
	EbsVolumeType          *string
}

func BlockDeviceMappingFromEC2(i *ec2.BlockDeviceMapping) (string, *BlockDeviceMapping) {
	o := &BlockDeviceMapping{}
	o.VirtualName = i.VirtualName
	if i.Ebs != nil {
		o.EbsDeleteOnTermination = i.Ebs.DeleteOnTermination
		o.EbsVolumeSize = i.Ebs.VolumeSize
		o.EbsVolumeType = i.Ebs.VolumeType
	}
	return aws.StringValue(i.DeviceName), o
}

func (i *BlockDeviceMapping) ToEC2(deviceName string) *ec2.BlockDeviceMapping {
	o := &ec2.BlockDeviceMapping{}
	o.DeviceName = aws.String(deviceName)
	o.VirtualName = i.VirtualName
	if i.EbsDeleteOnTermination != nil || i.EbsVolumeSize != nil || i.EbsVolumeType != nil {
		o.Ebs = &ec2.EbsBlockDevice{}
		o.Ebs.DeleteOnTermination = i.EbsDeleteOnTermination
		o.Ebs.VolumeSize = i.EbsVolumeSize
		o.Ebs.VolumeType = i.EbsVolumeType
	}
	return o
}

func BlockDeviceMappingFromAutoscaling(i *autoscaling.BlockDeviceMapping) (string, *BlockDeviceMapping) {
	o := &BlockDeviceMapping{}
	o.VirtualName = i.VirtualName
	if i.Ebs != nil {
		o.EbsDeleteOnTermination = i.Ebs.DeleteOnTermination
		o.EbsVolumeSize = i.Ebs.VolumeSize
		o.EbsVolumeType = i.Ebs.VolumeType
	}
	return aws.StringValue(i.DeviceName), o
}

func (i *BlockDeviceMapping) ToAutoscaling(deviceName string) *autoscaling.BlockDeviceMapping {
	o := &autoscaling.BlockDeviceMapping{}
	o.DeviceName = aws.String(deviceName)
	o.VirtualName = i.VirtualName

	if i.EbsDeleteOnTermination != nil || i.EbsVolumeSize != nil || i.EbsVolumeType != nil {
		o.Ebs = &autoscaling.Ebs{}
		o.Ebs.DeleteOnTermination = i.EbsDeleteOnTermination
		o.Ebs.VolumeSize = i.EbsVolumeSize
		o.Ebs.VolumeType = i.EbsVolumeType
	}

	return o
}

var _ fi.HasDependencies = &BlockDeviceMapping{}

func (f *BlockDeviceMapping) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	return nil
}
