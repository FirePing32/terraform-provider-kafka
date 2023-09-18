if test -f "$HOME/.kafka/kafka.tgz"
then
    cd $HOME/.kafka
    mkdir kafka
    tar -xzf kafka.tgz -C $HOME/.kafka/kafka
else
    echo "kafka binary not found in $HOME/.kafka"
    exit 1
fi