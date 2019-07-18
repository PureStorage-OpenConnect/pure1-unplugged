# GUI
The GUI for Pure1 Unplugged powered by Angular.

## Building and testing
The `make` commands from the parent directory should be used to lint, build, and test the web content.
```
make lint-web-content
make web-content
make test-web-content
```
Additionally, `make web` will rebuild dependencies and do all of the above tasks as well.

## Modules
There are six modules used.

* `app-routing`: handles all of the routing for the app
* `core`: houses components that are used dispersed throughout the app, including the sidebar and "main page" / navbar
* `dashboard`: houses components not powered by any services, but that embed Kibana dashboards
* `device`: for the appliances page that displays card views and latest metrics
* `device-alert`: for the messages page that displays alerts
* `support`: for the support page that displays error and timer logs

## Shared resources
These resources (interfaces, models, and services) are not coupled to one individual module and similar resources should be stored together. The `mocks` are also to be used for all component and service tests.

## Making changes
All new features and bug fixes should be accompanied with new test cases.
Changes should be accompanied with a screenshot and will be linted and tested.
