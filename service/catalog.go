package service

import (
	"encoding/json"

	"github.com/sacloud/open-service-broker-sacloud/osb"
)

var (
	// CurrentCatalog is current available catalog object
	CurrentCatalog = &osb.Catalog{
		Services: []*osb.Service{
			MariaDBService,
			PostgreSQLService,
		},
	}
	// CurrentCatalogData is raw data of API server response
	CurrentCatalogData []byte
)

const (
	// MariaDBServiceID service/mariadb/id
	MariaDBServiceID = "c4f353f3-8a59-437d-b4af-6a6f856248db"

	// MariaDBPlan10GID plan/mariadb/10g/id
	MariaDBPlan10GID = "7320351e-4664-4df5-938c-c0bfe8551609"

	// MariaDBPlan30GID plan/mariadb/30g/id
	MariaDBPlan30GID = "d4591cca-6361-4957-bfad-3f5e84ef2215"

	// MariaDBPlan90GID plan/mariadb/90g/id
	MariaDBPlan90GID = "66907d28-6440-41ca-9f7f-ab9c4a0d171b"

	// MariaDBPlan240GID plan/mariadb/240g/id
	MariaDBPlan240GID = "b3321210-e5f0-4224-99cf-34abc4435ed3"

	// MariaDBPlan500GID plan/mariadb/500g/id
	MariaDBPlan500GID = "92a30279-e938-4e20-8e4e-50a0231e719e"

	// MariaDBPlan1TID plan/mariadb/1tb/id
	MariaDBPlan1TID = "8705f035-805b-4009-9158-1dd48aebcb41"

	// PostgreSQLServiceID service/postgres/id
	PostgreSQLServiceID = "cc17eabf-0178-4eea-966e-2f5fb2aa62a9"

	// PostgreSQLPlan10GID plan/mariadb/10g/id
	PostgreSQLPlan10GID = "590eb4a9-6efb-4f14-ac03-b66fe49adb2e"

	// PostgreSQLPlan30GID plan/mariadb/30g/id
	PostgreSQLPlan30GID = "feb02498-4cdb-404a-8606-4e56ec019d39"

	// PostgreSQLPlan90GID plan/mariadb/90g/id
	PostgreSQLPlan90GID = "bfff8edd-fb74-4e63-85e7-04f143db0a5f"

	// PostgreSQLPlan240GID plan/mariadb/240g/id
	PostgreSQLPlan240GID = "8ebb52e3-7b6a-47b6-87b3-cfe9af9b78bf"

	// PostgreSQLPlan500GID plan/mariadb/500g/id
	PostgreSQLPlan500GID = "64d02933-b8c4-45fc-afda-7e7045af5b3f"

	// PostgreSQLPlan1TID plan/mariadb/1tb/id
	PostgreSQLPlan1TID = "f4bf204d-a73f-4b41-b8e8-1417c47c18dc"

	databaseApplianceParameterJSON = `
    {
    	"$schema": "http://json-schema.org/draft-04/schema#",
        "properties": {
            "allowNetworks": {
                "items": {
                    "type": "string"
                },
                "type": "array"
            },
            "backupTime": {
                "type": "string"
            },
            "defaultRoute": {
                "type": "string"
            },
            "ipaddress": {
                "type": "string"
            },
            "maskLen": {
                "type": "integer"
            },
            "port": {
                "type": "integer"
            },
            "switchID": {
                "type": "integer"
            },
            "username": {
                "type": "string"
            }
        },
        "additionalProperties": false,
        "type": "object"
	}
    `
)

// DatabaseIDMap defines relations of between service and plans
var DatabaseIDMap = map[string]PlanIDMap{
	"MariaDB": {
		ID: MariaDBServiceID,
		PlanIDMap: map[int]string{
			10:   MariaDBPlan10GID,
			30:   MariaDBPlan30GID,
			90:   MariaDBPlan90GID,
			240:  MariaDBPlan240GID,
			500:  MariaDBPlan500GID,
			1000: MariaDBPlan1TID,
		},
	},
	"postgres": {
		ID: PostgreSQLServiceID,
		PlanIDMap: map[int]string{
			10:   PostgreSQLPlan10GID,
			30:   PostgreSQLPlan30GID,
			90:   PostgreSQLPlan90GID,
			240:  PostgreSQLPlan240GID,
			500:  PostgreSQLPlan500GID,
			1000: PostgreSQLPlan1TID,
		},
	},
}

var (
	// MariaDBService is service for manage to SAKURA cloud Database Appliances
	MariaDBService = &osb.Service{
		ID:             MariaDBServiceID,
		Name:           "sacloud-mariadb",
		Bindable:       true,
		PlanUpdateable: false,
		Tags:           []string{"database", "mariadb"},
		Description:    "SAKURA Cloud Database appliance(MariaDB)",
		Requires:       []string{},
		Metadata:       &osb.Metadata{},
		Plans: []*osb.Plan{
			MariaDBPlan10G,
			MariaDBPlan30G,
			MariaDBPlan90G,
			MariaDBPlan240G,
			MariaDBPlan500G,
			MariaDBPlan1T,
		},
	}

	// PostgreSQLService is service for manage to SAKURA cloud Database Appliances
	PostgreSQLService = &osb.Service{
		ID:             PostgreSQLServiceID,
		Name:           "sacloud-postgres",
		Bindable:       true,
		PlanUpdateable: false,
		Tags:           []string{"database", "postgres"},
		Description:    "SAKURA Cloud Database appliance(PostgreSQL)",
		Requires:       []string{},
		Metadata:       &osb.Metadata{},
		Plans: []*osb.Plan{
			PostgreSQLPlan10G,
			PostgreSQLPlan30G,
			PostgreSQLPlan90G,
			PostgreSQLPlan240G,
			PostgreSQLPlan500G,
			PostgreSQLPlan1T,
		},
	}
)

var (
	// MariaDBPlan10G is represents MariaDB 10g plan
	MariaDBPlan10G = &osb.Plan{
		ID:          MariaDBPlan10GID,
		Name:        "db-10g",
		Description: "DB 10GB",
		Bindable:    true,
		Free:        false,
		Metadata:    &osb.Metadata{},
		Schemas: &osb.SchemasObject{
			ServiceInstance: &osb.ServiceInstanceSchemaObject{
				Create: &osb.SchemaParameters{},
			},
			ServiceBinding: &osb.ServiceBindingSchemaObject{
				Create: &osb.SchemaParameters{},
			},
		},
	}
	// MariaDBPlan30G is represents MariaDB 30g plan
	MariaDBPlan30G = &osb.Plan{
		ID:          MariaDBPlan30GID,
		Name:        "db-30g",
		Description: "DB 30GB",
		Bindable:    true,
		Free:        false,
		Metadata:    &osb.Metadata{},
		Schemas: &osb.SchemasObject{
			ServiceInstance: &osb.ServiceInstanceSchemaObject{
				Create: &osb.SchemaParameters{},
			},
			ServiceBinding: &osb.ServiceBindingSchemaObject{
				Create: &osb.SchemaParameters{},
			},
		},
	}
	// MariaDBPlan90G is represents MariaDB 90g plan
	MariaDBPlan90G = &osb.Plan{
		ID:          MariaDBPlan90GID,
		Name:        "db-90g",
		Description: "DB 90GB",
		Bindable:    true,
		Free:        false,
		Metadata:    &osb.Metadata{},
		Schemas: &osb.SchemasObject{
			ServiceInstance: &osb.ServiceInstanceSchemaObject{
				Create: &osb.SchemaParameters{},
			},
			ServiceBinding: &osb.ServiceBindingSchemaObject{
				Create: &osb.SchemaParameters{},
			},
		},
	}
	// MariaDBPlan240G is represents MariaDB 240g plan
	MariaDBPlan240G = &osb.Plan{
		ID:          MariaDBPlan240GID,
		Name:        "db-240g",
		Description: "DB 240GB",
		Bindable:    true,
		Free:        false,
		Metadata:    &osb.Metadata{},
		Schemas: &osb.SchemasObject{
			ServiceInstance: &osb.ServiceInstanceSchemaObject{
				Create: &osb.SchemaParameters{},
			},
			ServiceBinding: &osb.ServiceBindingSchemaObject{
				Create: &osb.SchemaParameters{},
			},
		},
	}
	// MariaDBPlan500G is represents MariaDB 500g plan
	MariaDBPlan500G = &osb.Plan{
		ID:          MariaDBPlan500GID,
		Name:        "db-500g",
		Description: "DB 500GB",
		Bindable:    true,
		Free:        false,
		Metadata:    &osb.Metadata{},
		Schemas: &osb.SchemasObject{
			ServiceInstance: &osb.ServiceInstanceSchemaObject{
				Create: &osb.SchemaParameters{},
			},
			ServiceBinding: &osb.ServiceBindingSchemaObject{
				Create: &osb.SchemaParameters{},
			},
		},
	}
	// MariaDBPlan1T is represents MariaDB 1TB plan
	MariaDBPlan1T = &osb.Plan{
		ID:          MariaDBPlan1TID,
		Name:        "db-1t",
		Description: "DB 1TB",
		Bindable:    true,
		Free:        false,
		Metadata:    &osb.Metadata{},
		Schemas: &osb.SchemasObject{
			ServiceInstance: &osb.ServiceInstanceSchemaObject{
				Create: &osb.SchemaParameters{},
			},
			ServiceBinding: &osb.ServiceBindingSchemaObject{
				Create: &osb.SchemaParameters{},
			},
		},
	}
)

var (
	// PostgreSQLPlan10G is represents PostgreSQL 10g plan
	PostgreSQLPlan10G = &osb.Plan{
		ID:          PostgreSQLPlan10GID,
		Name:        "db-10g",
		Description: "DB 10GB",
		Bindable:    true,
		Free:        false,
		Metadata:    &osb.Metadata{},
		Schemas: &osb.SchemasObject{
			ServiceInstance: &osb.ServiceInstanceSchemaObject{
				Create: &osb.SchemaParameters{},
			},
			ServiceBinding: &osb.ServiceBindingSchemaObject{
				Create: &osb.SchemaParameters{},
			},
		},
	}
	// PostgreSQLPlan30G is represents PostgreSQL 30g plan
	PostgreSQLPlan30G = &osb.Plan{
		ID:          PostgreSQLPlan30GID,
		Name:        "db-30g",
		Description: "DB 30GB",
		Bindable:    true,
		Free:        false,
		Metadata:    &osb.Metadata{},
		Schemas: &osb.SchemasObject{
			ServiceInstance: &osb.ServiceInstanceSchemaObject{
				Create: &osb.SchemaParameters{},
			},
			ServiceBinding: &osb.ServiceBindingSchemaObject{
				Create: &osb.SchemaParameters{},
			},
		},
	}
	// PostgreSQLPlan90G is represents PostgreSQL 90g plan
	PostgreSQLPlan90G = &osb.Plan{
		ID:          PostgreSQLPlan90GID,
		Name:        "db-90g",
		Description: "DB 90GB",
		Bindable:    true,
		Free:        false,
		Metadata:    &osb.Metadata{},
		Schemas: &osb.SchemasObject{
			ServiceInstance: &osb.ServiceInstanceSchemaObject{
				Create: &osb.SchemaParameters{},
			},
			ServiceBinding: &osb.ServiceBindingSchemaObject{
				Create: &osb.SchemaParameters{},
			},
		},
	}
	// PostgreSQLPlan240G is represents PostgreSQL 240g plan
	PostgreSQLPlan240G = &osb.Plan{
		ID:          PostgreSQLPlan240GID,
		Name:        "db-240g",
		Description: "DB 240GB",
		Bindable:    true,
		Free:        false,
		Metadata:    &osb.Metadata{},
		Schemas: &osb.SchemasObject{
			ServiceInstance: &osb.ServiceInstanceSchemaObject{
				Create: &osb.SchemaParameters{},
			},
			ServiceBinding: &osb.ServiceBindingSchemaObject{
				Create: &osb.SchemaParameters{},
			},
		},
	}
	// PostgreSQLPlan500G is represents PostgreSQL 500g plan
	PostgreSQLPlan500G = &osb.Plan{
		ID:          PostgreSQLPlan500GID,
		Name:        "db-500g",
		Description: "DB 500GB",
		Bindable:    true,
		Free:        false,
		Metadata:    &osb.Metadata{},
		Schemas: &osb.SchemasObject{
			ServiceInstance: &osb.ServiceInstanceSchemaObject{
				Create: &osb.SchemaParameters{},
			},
			ServiceBinding: &osb.ServiceBindingSchemaObject{
				Create: &osb.SchemaParameters{},
			},
		},
	}
	// PostgreSQLPlan1T is represents PostgreSQL 1TB plan
	PostgreSQLPlan1T = &osb.Plan{
		ID:          PostgreSQLPlan1TID,
		Name:        "db-1t",
		Description: "DB 1TB",
		Bindable:    true,
		Free:        false,
		Metadata:    &osb.Metadata{},
		Schemas: &osb.SchemasObject{
			ServiceInstance: &osb.ServiceInstanceSchemaObject{
				Create: &osb.SchemaParameters{},
			},
			ServiceBinding: &osb.ServiceBindingSchemaObject{
				Create: &osb.SchemaParameters{},
			},
		},
	}
)

func init() {
	var dbParamSchema map[string]interface{}

	err := json.Unmarshal([]byte(databaseApplianceParameterJSON), &dbParamSchema)
	if err != nil {
		panic(err)
	}
	MariaDBPlan10G.Schemas.ServiceInstance.Create.Parameters = dbParamSchema
	MariaDBPlan30G.Schemas.ServiceInstance.Create.Parameters = dbParamSchema
	MariaDBPlan90G.Schemas.ServiceInstance.Create.Parameters = dbParamSchema
	MariaDBPlan240G.Schemas.ServiceInstance.Create.Parameters = dbParamSchema
	MariaDBPlan500G.Schemas.ServiceInstance.Create.Parameters = dbParamSchema
	MariaDBPlan1T.Schemas.ServiceInstance.Create.Parameters = dbParamSchema

	PostgreSQLPlan10G.Schemas.ServiceInstance.Create.Parameters = dbParamSchema
	PostgreSQLPlan30G.Schemas.ServiceInstance.Create.Parameters = dbParamSchema
	PostgreSQLPlan90G.Schemas.ServiceInstance.Create.Parameters = dbParamSchema
	PostgreSQLPlan240G.Schemas.ServiceInstance.Create.Parameters = dbParamSchema
	PostgreSQLPlan500G.Schemas.ServiceInstance.Create.Parameters = dbParamSchema
	PostgreSQLPlan1T.Schemas.ServiceInstance.Create.Parameters = dbParamSchema
}

// PlanIDMap defines relations of between actual plan_id and osb plan_id
type PlanIDMap struct {
	ID        string
	PlanIDMap map[int]string
}
