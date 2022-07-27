# **Inventa Examples**

[![LinkedIn](https://img.shields.io/badge/LinkedIn-0077B5?style=for-the-badge&logo=linkedin&logoColor=white&style=flat-square)](https://www.linkedin.com/in/alper-dalkiran/)
[![Twitter](https://img.shields.io/badge/Twitter-1DA1F2?style=for-the-badge&logo=twitter&logoColor=white&style=flat-square)](https://twitter.com/aalperdalkiran)
![HitCount](https://hits.dwyl.com/adalkiran/inventa-examples.svg?style=flat-square)
![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)

Cross-language example projects to demonstrate how Inventa works and how to use it as microservice registry and for executing RPC.

## **WHY THIS PROJECT?**

This project aims to present you how to use the Inventa library. Inventa can be used in Orchestrator or Service roles. You can find example usages for both roles. E.g., the calculator/go-orchestrator project uses Inventa on Orchestrator role. The other projects use Inventa on Service role.

For more information, play with the project in "calculator" folder, as follows:

## **INSTALLATION and RUNNING**

This project was designed to run in Docker Container. Docker Compose file creates some containers, with some replica instances:
* **redis:** Runs a Redis instance.
* **go-orchestrator:** The only orchestrator in the calculator project. Other services will register themselves to this application. After this, it will call:
    * "calculate-sum" and "calculate-subtract" remote procedures with two random integer arguments, in every 2 seconds.
        <br>
        For every one of these calls, one of the "go-calculator-service" and "py-calculator-service" instances is selected by the orchestrator (they were registered to it), then the procedure is called specifically on the selected instance. You can see which programming language developed the service with, in the orchestrator's logs.
    * "linalg-matmul" remote procedure with different arguments, in every 3 seconds. This method will be available only in Pythonic service (py-linalg-service) which uses [NumPy](https://numpy.org) library.
        <br>
        Also, this example shows how we can transfer matrices between Go and Python services, via Redis. It will convert each cell value to 4-bytes, encoding as big-endian style.
        * Arguments will be two random matrices, which are suitable for matrix multiplication. To be valid, matrices should be in shapes like A=(m x n), B=(n, p). This first call shows how a valid response is generated.
        * Arguments will be two constant matrices, which are suitable for matrix multiplication.
        * Arguments will be two constant matrices, which are NOT suitable for matrix multiplication. This example will show you how the service will respond you the error message.
    
* **go-calculator-service:** The service will register itself to the orchestrator, and can respond to "calculate-sum" and "calculate-subtract" procedure calls. Written in Go language.
    <br>
    Can be more than one, by docker-compose.yml file's replica values, default is 5.
* **py-calculator-service:** The service will register itself to the orchestrator, and can respond to "calculate-sum" and "calculate-subtract" procedure calls. Functioning the same as go-calculator-service. Written in Python language.
    <br>
    Can be more than one, by docker-compose.yml file's replica values, default is 5.
* **py-linalg-service:** The service will register itself to the orchestrator, and can respond to "linalg-matmul" procedure calls. Written in Python language. It can do matrix multiplication, using [NumPy](https://numpy.org) library.
    <br>
    Can be more than one, by docker-compose.yml file's replica values, default is 5.

You can run it in production mode or development mode.

### **Production Mode**

* Clone this repo and run in terminal:

```sh
$ cd calculator
$ docker-compose up -d
```

* Wait until Go and Python modules were installed and configured. This can take some time. You can check out the download status by:

```sh
$ docker-compose logs -f
```

* After waiting for enough time, you will see the results of containers. If you are only interested in the result outputs of calculation services, you can only track the orchestrator's logs by:

```sh
$ docker logs -f calculator-go-orchestrator-1
```

### <a name="dev-mode"></a>**Development Mode: VS Code Remote - Containers**

To continue with VS Code and if this is your first time to work with Remote Containers in VS Code, you can check out [this link](https://code.visualstudio.com/docs/remote/containers) to learn how Remote Containers work in VS Code and follow the installation steps of [Remote Development extension pack](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.vscode-remote-extensionpack).

Then, follow these steps:

* Clone this repo to your local filesystem
* Open the folder "inventa-examples" with VS Code by "Open Folder..." command. This opens the root folder of the project.
* Press <kbd>F1</kbd> and select **"Remote Containers: Open Folder in Container..."** then select one of folders in the "calculator" folder, not the "calculator" folder itself. You can select any of the services, which you want to develop.
* This command creates (if they don't exist) required containers in Docker, then connects inside of calculator-[your-selected-service] container for development and debugging purposes.
* Wait until the containers are created, configured, and related VS Code server extensions installed inside the container. This can take some time. VS Code can ask for some required installations, click "Install All" for these prompts.
* After completion of all installations, press <kbd>F5</kbd> to start server application.
<br>
**Note:** Maybe you must kill existing running service processes by terminal.
* Then, you can keep track other services with docker logs.

<br>

## **INVENTA IMPLEMENTATIONS**

* Go implementation of Inventa on [Inventa for Go (go-inventa)](https://github.com/adalkiran/go-inventa).

* Python implementation of Inventa on [Inventa for Python (py-inventa)](https://github.com/adalkiran/py-inventa).

## **LICENSE**

Inventa Examples project is licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE) for the full license text.