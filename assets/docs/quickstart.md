# MUI Server Development Quick Start

The repository includes [front-end](#front-end) and [back-end](#back-end) programs.

## Front End

### Front End Development Requires

- Vue framework
- Typescript

### Build UI Static Files For Production

- Inside the server options, you need to set the `ui-dir` option to serve ui static files directory.

  To generate the ui static files, you should go under [public](../../ui/public) package.

  First, create an environment file called `config.js`.

  ```text
  window.config = {
        ApiUrl: '{{YOUR_REST_API_ADDRESS}}'
  }
  ```

### How To Develop Front End

The following example uses carrier maintenance files to illustrate

### Directory Structure For Quick Development

```txt
mui
└── ui                               # ui files
    └── src                          # main source code
        |── api                      # api service
        |── lang                     # i18n language
        |── views                    # views
        └── router                   # router

```

#### 1. Src

main source code.

#### 2. Api

```text
src
└── api                      #api service
    |── carrier.ts           #API request and interface definition corresponding to Swagger file
    └── carrierTypes.d.ts    #Interface definition
```

##### API Description

- `carrier.ts` is main to write carrier maintenance API, you can refer to the writing in the file.

    The default value defined by the interface will also be written here.

- `carrierTypes.d.ts` is the interface definition of carrier maintenance, which can define the required data types.

    For example, the return format of the API, the preset format of the table, etc.

#### 3. Lang

```text
src
└── lang                      # api service
    |── tw.ts                 # Traditional Chinese profile
    |── cn.ts                 # Simplified Chinese profile
    |── en.ts                 # English profile
    └── index.ts              # Language profile
```

##### Lang Description

- [tw.ts](../../ui/src/lang/tw.ts)

    To set the traditional Chinese translation, please set the index and corresponding text according to the file setting method.

- [cn.ts](../../ui/src/lang/cn.ts)

    To set the place for Simplified Chinese translation, please set the index and corresponding text according to the file setting method.

- [en.ts](../../ui/src/lang/en.ts)

    To set the place for English translation, please set the index and corresponding text according to the file setting method.

> The above three files must be added synchronously as long as the text to be displayed on the UI.

- [index.ts](../../ui/src/lang/index.ts)

    Places set for all languages, if you want to add other languages, you need to add them here.

#### 4. Views

```text
src
└── views                                 # views
    |── dashboard                         # dashboard
    |── login                             # login
    └── mesSystem                         # mes system
        └── carrierMaintenance            # carrier maintenance
            |── carrierMaintenance.css
            |── carrierMaintenance.ts
            └── carrierMaintenance.vue
```

##### Views Description

At present, the main developed system pages will be in the mesSystem folder, such as: machine maintenance page, inventory movement, etc. In order to facilitate the management of data, the code data is divided into .css, .ts, .vue and other files. The file naming method is lower camel case.

- `.css` is mainly used to manage the screen style of the page, if you need to set the style style, you can plan here.

- `.ts` is mainly used for logic control, data sorting, receiving data, etc. It belongs to the part of the brain, and it is also the place where we mainly need to write and sort.

- `.vue` is mainly used to lay out the UI page, and receive the data organized by .ts and render it on the screen for users to browse and read.

#### 5. Router

```text
src
└── router                      # router
```

##### Router Description

Mainly to set the path of the view. When the path is set, you can link this path to the desired page on other pages.

## Steps To Create A UI Page

1. According to the swagger document, create the corresponding file in the [api](../../ui/src/api) folder.

2. According to the UI draft or user needs, create corresponding data in [lang](../../ui/src/lang) for the text that is not in the language file.

3. Then create a page, if it is a system page, please create it in [mesSystem](../../ui/src/views/mesSystem), if it is for other functions, you can create it in [views](../../ Create a folder under the path of ui/src/views), and then create a file from the folder.

4. Finally, set the page path just created in [router](../../ui/src/router.ts).

### Front End Merge Request Example

[Carrier Maintenance UI](https://gitlab.kenda.com.tw/kenda/mui/-/merge_requests/88)

## Back End

### Back End Development Requires

- golang
- Protocol Buffers
- OpenAPI 2.0

### Error Handlers

- 400 (Bad Request): occurs when there's a chance to tell user to fix the bad request.
- 401 (Unauthorized): occurs when validate user token.
- 403 (Forbidden): occurs when user has no permission to access handler.
- 408 (Request Timeout): occurs when the request exceeds.
- 500 (Internal Server Error): occurs when server has internal error.

### How To Develop Back End

#### 1. Write Swagger API Document

We are using OpenAPI Specification 2.0 on MUI project for RESTful API Standardization which located under our project path, named: **swagger.yml**.

Usually, We classify each API features by using tags, for example: you want to create user login and logout features, you should classify them under _user_ tag.

OpenAPI Specification 2.0 Document: [OAS 2.0](https://swagger.io/specification/v2/).

Preview Swagger: [Swagger Editor](https://editor.swagger.io/).

#### 2. Generate Swagger

As you can see, there's no Golang main package and API components available under the [server](./server) directory, to generate them you can follow [generator](./server/swagger_gen.go) for swagger generation.

**IMPORTANT NOTE FOR DEVELOPER:**
if you are going to change any definition on [swagger.yml](./swagger.yml) , you should follow remove the generated directory first (listed inside [.gitignore](./server/swagger/.gitignore) file) before regenerate swagger server.

#### 3. Implementations and Unit-tests

After generated the code, there're mandatory steps you must follow:

1. Create New Features Function Handler Name on [func.proto](../protobuf/kenda/func.proto).

   Naming Style: all capital words separated with underscore; e.g: LIST_CARRIER

2. Create Services.

   located in: [service.go](../../server/impl/service/service.go)

   Basically we classify the API service by swagger tag, for example: you've created a new tag named: 'Carrier', then inside [service.go](../../server/impl/service/service.go), you need to create carrier service interface.

   ```go
   // Carrier service available function methods.
   type Carrier interface {
      // API handler list
   }
   ```

   Inside the interface, you must declare the API handler list which this service offers.

   ```go
   // Carrier service available function methods.
   type Carrier interface {
         List(params carrier.GetCarrierListParams, principal *models.Principal) middleware.Responder
         Create(params carrier.CreateCarrierParams, principal*models.Principal) middleware.Responder
         Update(params carrier.UpdateCarrierParams, principal *models.Principal) middleware.Responder
         Delete(params carrier.DeleteCarrierParams, principal*models.Principal) middleware.Responder
   }
   ```

   In this case, we got `List`, `Create`, `Update` and `Delete` methods for Carrier Service, then you must register the `Carrier interface` inside `Service struct` and the corresponding New Service registration.

3. Register Handlers.

   Located inside: [register.go](../../server/impl/handlers/mcom/register.go)

   Inside `RegisterHandlers` function, you assign the API handler function.

   ```go
   // Carrier handlers.
    api.CarrierGetCarrierListHandler = carrier.GetCarrierListHandlerFunc(s.Carrier().List)

    api.CarrierCreateCarrierHandler = carrier.CreateCarrierHandlerFunc(s.Carrier().Create)

    api.CarrierUpdateCarrierHandler = carrier.UpdateCarrierHandlerFunc(s.Carrier().Update)

    api.CarrierDeleteCarrierHandler = carrier.DeleteCarrierHandlerFunc(s.Carrier().Delete)
   ```

4. Implement your service and handlers.

   you start implementing your API inside [handlers](../../server/impl/handlers/mcom/) path.
   Basically, first classify the file package by tags, for instance: you have created 'carrier' tag which included `List`, `Create`, `Update` and `Delete` handler function, then inside [handlers](../../server/impl/handlers/mcom/) path, you need to create a new directory: **carrier** and inside carrier file directory, you will create your handler implementations file call `impl.go` (where you will write the handler functions implementation) and unit-test file call `impl_test.go` (where you test cases for your handler function).

#### 4. Run/Build Server

You can build server under the [server](./server) directory.

But before you run this server, you should set some configuration options to build the server.

First, you should find the Golang `main` package, it will be located inside `./server/swagger/${your- swagger-server-name}-server/main.go`

After you found the Golang main package, you can list out the configuration options by typing:

  ```console
  go run -h
  ```

Example of serves RESTful API at :8888

  ```console
  go run swagger/cmd/mui-server/main.go --scheme=http --host=0.0.0.0 --port=8888 --server-config=assets/.ignore/configs.yaml
  ```

In order to send the request to the station, please download mesage [openapi.yaml]`https://gitlab.kenda.com.tw/kenda/mesage/-/blob/${branch}/openapi.yaml`, put into `./assets/mesage`(in which ${branch} inside the URL reference is the specified branch name of the corresponding features)

### Back End Merge Request Example

[Carrier Maintenance API](https://gitlab.kenda.com.tw/kenda/mui/-/merge_requests/84)
