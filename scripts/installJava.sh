if type -p java; then
    echo found java executable in PATH
    _java=java
elif [[ -n "$JAVA_HOME" ]] && [[ -x "$JAVA_HOME/bin/java" ]];  then
    echo found java executable in JAVA_HOME
    _java="$JAVA_HOME/bin/java"
else
    echo no java executable is installed. Installing...
    sudo apt update
    sudo apt install default-jre
    sudo apt install default-jdk
    echo installed latest version of Java JRE and JDK
    exit 0
fi