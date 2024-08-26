#!/bin/sh

# проверяем наличие нужных для сервера тулз
check_and_install_requirements() {
    if command -v $1 >/dev/null 2>&1; then
        echo "$1 is already installed."
        return 0
    else
        echo "$1 is not installed. Starting install..."
        install_requirements $1
        return 1
    fi
}

# устанавливаем нужные тулзы
install_requirements() {
    OS_TYPE=$(uname)

    case "$OS_TYPE" in
        Linux*)
            echo "Detected Linux. Checking package manager..."
            if command -v apt-get >/dev/null 2>&1; then
                echo "Using apt-get to install $1."
                sudo apt-get update
                sudo apt-get install -y $1
            elif command -v yum >/dev/null 2>&1; then
                echo "Using yum to install $1."
                sudo yum install -y $1
            elif command -v dnf >/dev/null 2>&1; then
                echo "Using dnf to install $1."
                sudo dnf install -y $1
            elif command -v zypper >/dev/null 2>&1; then
                echo "Using zypper to install $1."
                sudo zypper install -y $1
            else
                echo "No supported package manager found. Please install $1 manually."
                exit 1
            fi
            ;;
        Darwin*)
            echo "Detected macOS. Using Homebrew to install $1."
            if command -v brew >/dev/null 2>&1; then
                brew install $1
            else
                echo "Homebrew is not installed. Please install Homebrew and then $1 manually."
                exit 1
            fi
            ;;
        *)
            echo "Unsupported OS: $OS_TYPE"
            exit 1
            ;;
    esac
}

# проверка нужных тулз для сервера
check_and_install_requirements "tail"
check_and_install_requirements "wc"
