//go:build (all || resource_serviceendpoint_maven) && !exclude_serviceendpoints
// +build all resource_serviceendpoint_maven
// +build !exclude_serviceendpoints

package serviceendpoint

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/microsoft/azure-devops-go-api/azuredevops/v6/serviceendpoint"
	"github.com/microsoft/terraform-provider-azuredevops/azdosdkmocks"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/internal/client"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/internal/utils/converter"
	"github.com/stretchr/testify/require"
)

var mavenTestServiceEndpointIDpassword = uuid.New()
var mavenRandomServiceEndpointProjectIDpassword = uuid.New()
var mavenTestServiceEndpointProjectIDpassword = &mavenRandomServiceEndpointProjectID

var mavenTestServiceEndpointPassword = serviceendpoint.ServiceEndpoint{
	Authorization: &serviceendpoint.EndpointAuthorization{
		Parameters: &map[string]string{
			"username": "MAVEN_TEST_USERNAME",
			"password": "MAVEN_TEST_PASSWORD",
		},
		Scheme: converter.String("UsernamePassword"),
	},
	Data: &map[string]string{
		"RepositoryId": "MAVEN_TESTrepo",
	},
	Id:    &mavenTestServiceEndpointIDpassword,
	Name:  converter.String("UNIT_TEST_CONN_NAME"),
	Owner: converter.String("library"), // Supported values are "library", "agentcloud"
	Type:  converter.String("externalmavenrepository"),
	Url:   converter.String("https://www.maven.com"),
	ServiceEndpointProjectReferences: &[]serviceendpoint.ServiceEndpointProjectReference{
		{
			ProjectReference: &serviceendpoint.ProjectReference{
				Id: mavenTestServiceEndpointProjectIDpassword,
			},
			Name:        converter.String("UNIT_TEST_CONN_NAME"),
			Description: converter.String("UNIT_TEST_CONN_DESCRIPTION"),
		},
	},
}

var mavenTestServiceEndpointID = uuid.New()
var mavenRandomServiceEndpointProjectID = uuid.New()
var mavenTestServiceEndpointProjectID = &mavenRandomServiceEndpointProjectID

var mavenTestServiceEndpoint = serviceendpoint.ServiceEndpoint{
	Authorization: &serviceendpoint.EndpointAuthorization{
		Parameters: &map[string]string{
			"apitoken": "MAVEN_TEST_TOKEN",
		},
		Scheme: converter.String("Token"),
	},
	Data: &map[string]string{
		"RepositoryId": "MAVEN_TEST_REPO",
	},
	Id:    &mavenTestServiceEndpointID,
	Name:  converter.String("UNIT_TEST_CONN_NAME"),
	Owner: converter.String("library"), // Supported values are "library", "agentcloud"
	Type:  converter.String("externalmavenrepository"),
	Url:   converter.String("https://www.maven.com"),
	// RepositoryId: converter.String("Test-Repo"),
	ServiceEndpointProjectReferences: &[]serviceendpoint.ServiceEndpointProjectReference{
		{
			ProjectReference: &serviceendpoint.ProjectReference{
				Id: mavenTestServiceEndpointProjectID,
			},
			Name:        converter.String("UNIT_TEST_CONN_NAME"),
			Description: converter.String("UNIT_TEST_CONN_DESCRIPTION"),
		},
	},
}

// verifies that the flatten/expand round trip yields the same service endpoint
func testServiceEndpointMaven_ExpandFlatten_Roundtrip(t *testing.T, ep *serviceendpoint.ServiceEndpoint, id *uuid.UUID) {
	for _, ep := range []*serviceendpoint.ServiceEndpoint{ep, ep} {
		resourceData := schema.TestResourceDataRaw(t, ResourceServiceEndpointMaven().Schema, nil)
		flattenServiceEndpointMaven(resourceData, ep, id)

		serviceEndpointAfterRoundTrip, projectID, err := expandServiceEndpointMaven(resourceData)

		require.Nil(t, err)
		require.Equal(t, *ep, *serviceEndpointAfterRoundTrip)
		require.Equal(t, id, projectID)
	}
}
func TestServiceEndpointMaven_ExpandFlatten_RoundtripPassword(t *testing.T) {
	testServiceEndpointMaven_ExpandFlatten_Roundtrip(t, &mavenTestServiceEndpointPassword, mavenTestServiceEndpointProjectIDpassword)
}

func TestServiceEndpointMaven_ExpandFlatten_RoundtripToken(t *testing.T) {
	testServiceEndpointMaven_ExpandFlatten_Roundtrip(t, &mavenTestServiceEndpoint, mavenTestServiceEndpointProjectID)
}

// verifies that if an error is produced on create, the error is not swallowed
func testServiceEndpointMaven_Create_DoesNotSwallowError(t *testing.T, ep *serviceendpoint.ServiceEndpoint, id *uuid.UUID) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r := ResourceServiceEndpointMaven()
	resourceData := schema.TestResourceDataRaw(t, r.Schema, nil)
	flattenServiceEndpointMaven(resourceData, ep, id)

	buildClient := azdosdkmocks.NewMockServiceendpointClient(ctrl)
	clients := &client.AggregatedClient{ServiceEndpointClient: buildClient, Ctx: context.Background()}

	expectedArgs := serviceendpoint.CreateServiceEndpointArgs{Endpoint: ep}
	buildClient.
		EXPECT().
		CreateServiceEndpoint(clients.Ctx, expectedArgs).
		Return(nil, errors.New("CreateServiceEndpoint() Failed")).
		Times(1)

	err := r.Create(resourceData, clients)
	require.Contains(t, err.Error(), "CreateServiceEndpoint() Failed")
}
func TestServiceEndpointMaven_Create_DoesNotSwallowErrorToken(t *testing.T) {
	testServiceEndpointMaven_Create_DoesNotSwallowError(t, &mavenTestServiceEndpoint, mavenTestServiceEndpointProjectID)
}
func TestServiceEndpointMaven_Create_DoesNotSwallowErrorPassword(t *testing.T) {
	testServiceEndpointMaven_Create_DoesNotSwallowError(t, &mavenTestServiceEndpointPassword, mavenTestServiceEndpointProjectIDpassword)
}

// verifies that if an error is produced on read, the error is not swallowed
func testServiceEndpointMaven_Read_DoesNotSwallowError(t *testing.T, ep *serviceendpoint.ServiceEndpoint, id *uuid.UUID) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r := ResourceServiceEndpointMaven()
	resourceData := schema.TestResourceDataRaw(t, r.Schema, nil)
	flattenServiceEndpointMaven(resourceData, ep, id)

	buildClient := azdosdkmocks.NewMockServiceendpointClient(ctrl)
	clients := &client.AggregatedClient{ServiceEndpointClient: buildClient, Ctx: context.Background()}

	expectedArgs := serviceendpoint.GetServiceEndpointDetailsArgs{
		EndpointId: ep.Id,
		Project:    converter.String(id.String()),
	}
	buildClient.
		EXPECT().
		GetServiceEndpointDetails(clients.Ctx, expectedArgs).
		Return(nil, errors.New("GetServiceEndpoint() Failed")).
		Times(1)

	err := r.Read(resourceData, clients)
	require.Contains(t, err.Error(), "GetServiceEndpoint() Failed")
}
func TestServiceEndpointMaven_Read_DoesNotSwallowErrorToken(t *testing.T) {
	testServiceEndpointMaven_Read_DoesNotSwallowError(t, &mavenTestServiceEndpoint, mavenTestServiceEndpointProjectID)
}
func TestServiceEndpointMaven_Read_DoesNotSwallowErrorPassword(t *testing.T) {
	testServiceEndpointMaven_Read_DoesNotSwallowError(t, &mavenTestServiceEndpointPassword, mavenTestServiceEndpointProjectIDpassword)
}

// verifies that if an error is produced on a delete, it is not swallowed
func testServiceEndpointMaven_Delete_DoesNotSwallowError(t *testing.T, ep *serviceendpoint.ServiceEndpoint, id *uuid.UUID) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r := ResourceServiceEndpointMaven()
	resourceData := schema.TestResourceDataRaw(t, r.Schema, nil)
	flattenServiceEndpointMaven(resourceData, ep, id)

	buildClient := azdosdkmocks.NewMockServiceendpointClient(ctrl)
	clients := &client.AggregatedClient{ServiceEndpointClient: buildClient, Ctx: context.Background()}

	expectedArgs := serviceendpoint.DeleteServiceEndpointArgs{
		EndpointId: ep.Id,
		ProjectIds: &[]string{
			id.String(),
		},
	}
	buildClient.
		EXPECT().
		DeleteServiceEndpoint(clients.Ctx, expectedArgs).
		Return(errors.New("DeleteServiceEndpoint() Failed")).
		Times(1)

	err := r.Delete(resourceData, clients)
	require.Contains(t, err.Error(), "DeleteServiceEndpoint() Failed")
}
func TestServiceEndpointMaven_Delete_DoesNotSwallowErrorToken(t *testing.T) {
	testServiceEndpointMaven_Delete_DoesNotSwallowError(t, &mavenTestServiceEndpoint, mavenTestServiceEndpointProjectID)
}
func TestServiceEndpointMaven_Delete_DoesNotSwallowErrorPassword(t *testing.T) {
	testServiceEndpointMaven_Delete_DoesNotSwallowError(t, &mavenTestServiceEndpointPassword, mavenTestServiceEndpointProjectIDpassword)
}

// verifies that if an error is produced on a update, it is not swallowed
func testServiceEndpointMaven_Update_DoesNotSwallowError(t *testing.T, ep *serviceendpoint.ServiceEndpoint, id *uuid.UUID) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	r := ResourceServiceEndpointMaven()
	resourceData := schema.TestResourceDataRaw(t, r.Schema, nil)
	flattenServiceEndpointMaven(resourceData, ep, id)

	buildClient := azdosdkmocks.NewMockServiceendpointClient(ctrl)
	clients := &client.AggregatedClient{ServiceEndpointClient: buildClient, Ctx: context.Background()}

	expectedArgs := serviceendpoint.UpdateServiceEndpointArgs{
		Endpoint:   ep,
		EndpointId: ep.Id,
	}

	buildClient.
		EXPECT().
		UpdateServiceEndpoint(clients.Ctx, expectedArgs).
		Return(nil, errors.New("UpdateServiceEndpoint() Failed")).
		Times(1)

	err := r.Update(resourceData, clients)
	require.Contains(t, err.Error(), "UpdateServiceEndpoint() Failed")
}
func TestServiceEndpointMaven_Update_DoesNotSwallowErrorToken(t *testing.T) {
	testServiceEndpointMaven_Delete_DoesNotSwallowError(t, &mavenTestServiceEndpoint, mavenTestServiceEndpointProjectID)
}
func TestServiceEndpointMaven_Update_DoesNotSwallowErrorPassword(t *testing.T) {
	testServiceEndpointMaven_Delete_DoesNotSwallowError(t, &mavenTestServiceEndpointPassword, mavenTestServiceEndpointProjectIDpassword)
}
