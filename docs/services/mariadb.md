# MariaDB - SAKURA Cloud database appliance

## Services & Plans

### Service: sacloud-mariadb

| Plan Name | Description |
|-----------|-------------|
| `db-10g`  | 10GB storage plan  |
| `db-30g`  | 30GB storage plan  |
| `db-90g`  | 90GB storage plan  |
| `db-240g` | 240GB storage plan |
| `db-500g` | 500GB storage plan |
| `db-1t`   | 1TB storage plan   |

#### Behaviors

##### Provision

Provisions a new MariaDB appliance instance.  
 
###### Provisioning Parameters

| Parameter Name | Type | Description | Required | Default Value |
|----------------|------|-------------|----------|---------------|
| `switchID` | `int64` | ID of the switch to which the database connects. | Required | Switch must be reachable from within the kubernetes cluster.|
| `ipaddress` | `string` | IP address to assign to the database. | Required | IP address must be reachable from within the kubernetes cluster. |
| `maskLen` | `int` | Network mask length to assign to the database. | Required | -|
| `defaultRoute` | `string` | Default route IP address to assign to the database. | Required | -|
| `port`          | `int` | The port number on which the database listens | N| `3306`|

##### Bind

Creates a new user and database on the MariaDB appliance.
The new user will be named randomly and will be granted a wide array of permissions on the database.
And the new database is created with the same name as the user name.

###### Binding Parameters

This binding operation does not support any parameters.

###### Credentials

Binding returns the following connection details and credentials:

| Field Name | Type | Description |
|------------|------|-------------|
| `host` | `string` | The fully-qualified address of the MariaDB DBMS. |
| `port` | `int` | The port number to connect to on the MariaDB DBMS. |
| `database` | `string` | The name of the database. |
| `username` | `string` | The name of the database user. |
| `password` | `string` | The password for the database user. |
| `sslRequired` | `boolean` | Flag indicating if SSL is required to connect the MariaDB DBMS. |
| `uri` | `string` | A URI string containing all necessary connection information. |

##### Unbind

Drops the applicable database and user from the MariaDB DBMS.

##### Deprovision

Deletes the MariaDB appliance.

##### Examples


The `examples/mariadb-service.yaml` can be used to provision the `sacloud-10g` plan.
This can be done with the following example:

```console
# Put your SAKURA Cloud resource settings to service instance definition
vi examples/mariadb-service.yaml

# create service
kubectl create -f examples/mariadb-service.yaml
```

You can then create a binding with the following command:

```console
kubectl create -f examples/mariadb-binding.yaml
```

