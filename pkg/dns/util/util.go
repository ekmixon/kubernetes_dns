/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package util

import (
	"fmt"
	"hash/fnv"
	"net"
	"strconv"
	"strings"

	"github.com/skynetservices/skydns/msg"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

const (
	// ArpaSuffix is the standard suffix for PTR IP reverse lookups.
	ArpaSuffix = ".in-addr.arpa."
	// defaultPriority used for service records
	defaultPriority = 10
	// defaultWeight used for service records
	defaultWeight = 10
	// defaultTTL used for service records
	defaultTTL = 30
)

// ExtractIP turns a standard PTR reverse record lookup name
// into an IP address
func ExtractIP(reverseName string) (string, bool) {
	if !strings.HasSuffix(reverseName, ArpaSuffix) {
		return "", false
	}
	search := strings.TrimSuffix(reverseName, ArpaSuffix)

	// reverse the segments and then combine them
	segments := ReverseArray(strings.Split(search, "."))
	return strings.Join(segments, "."), true
}

// ReverseArray reverses an array.
func ReverseArray(arr []string) []string {
	for i := 0; i < len(arr)/2; i++ {
		j := len(arr) - i - 1
		arr[i], arr[j] = arr[j], arr[i]
	}
	return arr
}

// Returns record in a format that SkyDNS understands.
// Also return the hash of the record.
func GetSkyMsg(ip string, port int) (*msg.Service, string) {
	msg := NewServiceRecord(ip, port)
	hash := HashServiceRecord(msg)
	klog.V(5).Infof("Constructed new DNS record: %s, hash:%s",
		fmt.Sprintf("%v", msg), hash)
	return msg, fmt.Sprintf("%x", hash)
}

// NewServiceRecord creates a new service DNS message.
func NewServiceRecord(ip string, port int) *msg.Service {
	return &msg.Service{
		Host:     ip,
		Port:     port,
		Priority: defaultPriority,
		Weight:   defaultWeight,
		Ttl:      defaultTTL,
	}
}

// HashServiceRecord hashes the string representation of a DNS
// message.
func HashServiceRecord(msg *msg.Service) string {
	s := fmt.Sprintf("%v", msg)
	h := fnv.New32a()
	h.Write([]byte(s))
	return fmt.Sprintf("%x", h.Sum32())
}

// ValidateNameserverIpAndPort splits and validates ip and port for nameserver.
// If there is no port in the given address, a default 53 port will be returned.
func ValidateNameserverIpAndPort(nameServer string) (string, string, error) {
	if ip := net.ParseIP(nameServer); ip != nil {
		return ip.String(), "53", nil
	}

	host, port, err := net.SplitHostPort(nameServer)
	if err != nil {
		return "", "", err
	}
	if ip := net.ParseIP(host); ip == nil {
		return "", "", fmt.Errorf("bad IP address: %q", host)
	}
	if p, err := strconv.Atoi(port); err != nil || p < 1 || p > 65535 {
		return "", "", fmt.Errorf("bad port number: %q", port)
	}
	return host, port, nil
}

// IsServiceIPSet aims to check if the service's ClusterIP is set or not
// the objective is not to perform validation here
func IsServiceIPSet(service *corev1.Service) bool {
	return service.Spec.ClusterIP != corev1.ClusterIPNone && service.Spec.ClusterIP != ""
}

// GetClusterIPs returns IPs set for the service
func GetClusterIPs(service *corev1.Service) []string {
	if len(service.Spec.ClusterIPs) > 0 {
		return service.Spec.ClusterIPs
	}
	return []string{service.Spec.ClusterIP}
}
