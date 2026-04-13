import QtQuick 2.0

Rectangle {
    id: main
    width: 400
    height: 300
    color: "blue"

    Text {
        id: textItem
        text: "Hello QML!"
        anchors.centerIn: parent
        color: "white"
    }

    MouseArea {
        anchors.fill: parent
        onClicked: {
            textItem.text = "Clicked!"
        }
    }
}
