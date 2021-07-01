package repoconfig

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("repoconfig parsing", func() {
	When("the yaml file has no label-approved", func() {
		It("should return no LabelsApproved", func() {
			config := []byte("")
			botConfig, err := parseConfigYAML(config)
			Expect(err).ToNot(HaveOccurred())
			Expect(botConfig.LabelApproved()).To(BeNil())
		})
	})

	When("the yaml file has label-approved only", func() {
		It("returns no label-approved", func() {
			config := []byte("label-approved:\n")
			botConfig, err := parseConfigYAML(config)
			Expect(err).ToNot(HaveOccurred())
			Expect(botConfig.LabelApproved()).To(BeNil())
		})
	})

	When("the yaml file has label-approved and label", func() {
		It("should return a valid object with the label", func() {
			config := []byte(`
label-approved:
  label: hey
`)
			botConfig, err := parseConfigYAML(config)
			Expect(err).ToNot(HaveOccurred())
			Expect(botConfig.LabelApproved()).ToNot(BeNil())
			Expect(botConfig.LabelApproved().Label()).To(Equal("hey"))
			Expect(botConfig.LabelApproved().Approvals()).To(Equal(defaultApprovals))
		})
	})

	When("the yaml file has label-approved and approvals", func() {
		It("should return a valid object with the approvals", func() {
			config := []byte(`
label-approved:
  approvals: 666
`)
			botConfig, err := parseConfigYAML(config)
			Expect(err).ToNot(HaveOccurred())
			Expect(botConfig.LabelApproved()).ToNot(BeNil())
			Expect(botConfig.LabelApproved().Label()).To(Equal(defaultLabel))
			Expect(botConfig.LabelApproved().Approvals()).To(Equal(666))
		})
	})

	When("the yaml file has a fully featured label-approved", func() {
		It("should return a valid object with all", func() {
			config := []byte(`
label-approved:
  label: hey
  approvals: 666
`)
			botConfig, err := parseConfigYAML(config)
			Expect(err).ToNot(HaveOccurred())
			Expect(botConfig.LabelApproved()).ToNot(BeNil())
			Expect(botConfig.LabelApproved().Label()).To(Equal("hey"))
			Expect(botConfig.LabelApproved().Approvals()).To(Equal(666))
		})
	})

	When("the yaml file is invalid", func() {
		It("should return error", func() {
			config := []byte("this is not valid")
			botConfig, err := parseConfigYAML(config)
			Expect(err).To(HaveOccurred())
			Expect(botConfig.LabelApproved()).To(BeNil())
		})
	})

})

func TestRepoconfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Repoconfig Suite")
}
