package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/sagemaker"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/sagemaker/finder"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func TestAccAWSSagemakerModelPackageGroupPolicy_basic(t *testing.T) {
	var mpg sagemaker.GetModelPackageGroupPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_model_package_group_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, sagemaker.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSSagemakerModelPackageGroupPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerModelPackageGroupPolicyBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerModelPackageGroupPolicyExists(resourceName, &mpg),
					resource.TestCheckResourceAttr(resourceName, "model_package_group_name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSSagemakerModelPackageGroupPolicy_disappears(t *testing.T) {
	var mpg sagemaker.GetModelPackageGroupPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_model_package_group_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, sagemaker.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSSagemakerModelPackageGroupPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerModelPackageGroupPolicyBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerModelPackageGroupPolicyExists(resourceName, &mpg),
					acctest.CheckResourceDisappears(acctest.Provider, ResourceModelPackageGroupPolicy(), resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, ResourceModelPackageGroupPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSSagemakerModelPackageGroupPolicy_disappears_modelPackageGroup(t *testing.T) {
	var mpg sagemaker.GetModelPackageGroupPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_sagemaker_model_package_group_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, sagemaker.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSSagemakerModelPackageGroupPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSagemakerModelPackageGroupPolicyBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSagemakerModelPackageGroupPolicyExists(resourceName, &mpg),
					acctest.CheckResourceDisappears(acctest.Provider, ResourceModelPackageGroup(), "aws_sagemaker_model_package_group.test"),
					acctest.CheckResourceDisappears(acctest.Provider, ResourceModelPackageGroupPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSSagemakerModelPackageGroupPolicyDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sagemaker_model_package_group_policy" {
			continue
		}

		_, err := finder.ModelPackageGroupPolicyByName(conn, rs.Primary.ID)
		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error reading Sagemaker Model Package Group Policy (%s): %w", rs.Primary.ID, err)
		}
	}

	return nil
}

func testAccCheckAWSSagemakerModelPackageGroupPolicyExists(n string, mpg *sagemaker.GetModelPackageGroupPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No sagmaker Model Package Group ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SageMakerConn
		resp, err := finder.ModelPackageGroupPolicyByName(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*mpg = *resp

		return nil
	}
}

func testAccAWSSagemakerModelPackageGroupPolicyBasicConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_iam_policy_document" "test" {
  statement {
    sid       = "AddPermModelPackageGroup"
    actions   = ["sagemaker:DescribeModelPackage", "sagemaker:ListModelPackages"]
    resources = [aws_sagemaker_model_package_group.test.arn]
    principals {
      identifiers = [data.aws_caller_identity.current.account_id]
      type        = "AWS"
    }
  }
}

resource "aws_sagemaker_model_package_group" "test" {
  model_package_group_name = %[1]q
}

resource "aws_sagemaker_model_package_group_policy" "test" {
  model_package_group_name = aws_sagemaker_model_package_group.test.model_package_group_name
  resource_policy          = jsonencode(jsondecode(data.aws_iam_policy_document.test.json))
}
`, rName)
}
