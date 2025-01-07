# TestYourServer

TestYourServer is a tool designed for server testing. It allows you to control the number of clients and the frequency of requests to stress test and analyze the performance of your server.

## 🚀 Installation & Building

### Prerequisites

Before building, make sure you have the following dependencies installed:

- **Windows:** MSYS2/MinGW64 and `make` are required.
- **Linux/macOS:** `make` should already be installed.

### 1. Install Dependencies

Run the following command to install the required dependencies:

#### For Linux/macOS:

```bash
make install-deps
```

#### For Windows (MSYS2/MinGW64):
Ensure you have MSYS2 or MinGW64 installed. Then, use the following command to install dependencies:

```bash
make install-deps
```

### 2. Build the Application
```bash
make
```

To build the project, you can choose between two modes:

Development Mode:

```bash
make build-dev
```

Production Mode:

```bash
make build-prod
```

### 3. Running the Application
Once the build is complete, you can run the application:

```bash
./build/TestYourServer
```
### 📝 Notes
Displaying Headers and Body of Requests: Enabling the display of request headers and bodies may cause lag, especially under heavy load, as visualizing the data requires additional resources.
Future Enhancements: We plan to implement the ability to modify request headers, use proxies, and dynamically update headers to make the tool even more versatile.
### 💻 Technologies Used
- Go for core functionality.
- Fyne for the graphical user interface.
- Make for project building and dependency management.
