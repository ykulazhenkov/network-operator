/*
Copyright 2021 NVIDIA

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

package state

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	mellanoxv1alpha1 "github.com/Mellanox/network-operator/api/v1alpha1"
	"github.com/Mellanox/network-operator/pkg/consts"
	"github.com/Mellanox/network-operator/pkg/render"
	"github.com/Mellanox/network-operator/pkg/testing/mocks"
	"github.com/Mellanox/network-operator/pkg/utils"
)

func checkRenderedDpCm(obj *unstructured.Unstructured, namespace, config string) {
	Expect(obj.GetKind()).To(Equal("ConfigMap"))
	Expect(obj.Object["metadata"].(map[string]interface{})["name"].(string)).To(Equal("sriovdp-config"))
	Expect(obj.Object["metadata"].(map[string]interface{})["namespace"].(string)).To(Equal(namespace))
	Expect(obj.Object["data"].(map[string]interface{})["config.json"].(string)).To(Equal(config))
}

func checkRenderedDpSA(obj *unstructured.Unstructured, namespace string) {
	Expect(obj.GetKind()).To(Equal("ServiceAccount"))
	Expect(obj.Object["metadata"].(map[string]interface{})["namespace"].(string)).To(Equal(namespace))
}

func checkRenderedDpDs(obj *unstructured.Unstructured, imageSpec *mellanoxv1alpha1.ImageSpec) {
	namespace := consts.NetworkOperatorResourceNamespace
	image := imageSpec.Repository + "/" + imageSpec.Image + ":" + imageSpec.Version
	template := obj.Object["spec"].(map[string]interface{})["template"].(map[string]interface{})
	spec := fmt.Sprintf("%v", template)

	Expect(obj.GetKind()).To(Equal("DaemonSet"))
	Expect(obj.Object["metadata"].(map[string]interface{})["namespace"].(string)).To(Equal(namespace))
	Expect(spec).To(ContainSubstring(image))
}

var _ = Describe("SR-IOV Device Plugin State tests", func() {

	Context("GetNodesAttributes with provide", func() {
		It("Should Apply", func() {
			client := mocks.ControllerRutimeClient{}
			manifestBaseDir := "../../manifests/stage-sriov-device-plugin"
			scheme := runtime.NewScheme()

			files, err := utils.GetFilesWithSuffix(manifestBaseDir, render.ManifestFileSuffix...)
			Expect(err).NotTo(HaveOccurred())
			renderer := render.NewRenderer(files)

			stateName := "state-SRIOV-device-plugin"
			sriovDpState := stateSriovDp{
				stateSkel: stateSkel{
					name:        stateName,
					description: "SR-IOV device plugin deployed in the cluster",
					client:      &client,
					scheme:      scheme,
					renderer:    renderer,
				},
			}

			Expect(err).NotTo(HaveOccurred())
			Expect(sriovDpState.Name()).To(Equal(stateName))

			cr := &mellanoxv1alpha1.NicClusterPolicy{}
			config := "config"

			imageSpec := &mellanoxv1alpha1.ImageSpec{
				Image:      "image",
				Repository: "Repository",
				Version:    "v0.0",
			}
			dpSpec := &mellanoxv1alpha1.DevicePluginSpec{
				ImageSpec: *imageSpec,
				Config:    config,
			}
			cr.Spec.SriovDevicePlugin = dpSpec
			objs, err := sriovDpState.getManifestObjects(cr)

			Expect(err).NotTo(HaveOccurred())
			Expect(len(objs)).To(Equal(3))

			namespace := consts.NetworkOperatorResourceNamespace

			checkRenderedDpCm(objs[0], namespace, config)
			checkRenderedDpSA(objs[1], namespace)
			checkRenderedDpDs(objs[2], imageSpec)
		})
	})
})
