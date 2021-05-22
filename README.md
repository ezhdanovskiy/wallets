# wallets

## Components
The application has two components:
1. postgres - database.
2. wallets - HTTP server.

### 1. postgres
The database contains two tables `wallets` and `operations`.  
- `wallets` contains wallet name and balance.
- `operations` contains deposit and withdrawal operations.  

Read `migrations` for details.  

I decided to use `sql.LevelSerializable` for transactions. 
This simplifies life on the stage of development.
In the future it will be possible to try to optimize.

### 2. wallets
The wallets component can be run in multiple instances.  
It serves four endpoints:
- `POST /wallets` - Add wallet.
- `POST /wallets/deposit` - Top up wallet.
- `POST /wallets/transfer` - Transfer money.
- `GET /wallets/operations` - Get wallet operations.

Read `api/v1/swagger.yaml` for details.

## Packages
- application - provides dependencies and runs application;
- config - reads configuration from envs (viper);
- consts - application constants;
- dto - Data Transfer Objects;
- http - HTTP server;
- httperr - custom errors;
- repository - client for DB (postgresql);
- service - business logic;
- tests - integration tests.

Use go doc for details.
```bash
go doc internal/service
```
