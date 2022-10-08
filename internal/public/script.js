const ws = new WebSocket("ws://localhost:8080/ws");

ws.onopen = ()=>{
    console.log("connection open");
    document.body.innerHTML="Connection open";
}

ws.onerror = (err)=>{
    console.log("error occured: ", err);
    document.body.innerHTML= err.message;
}

ws.onclose = () => {
    console.log("connection closed");
    document.body.innerHTML= "connection closed"
}


ws.onmessage = (event)=> {
    node  = document.createElement("pre")
    node.innerHTML = JSON.stringify(JSON.parse(event.data), null, 2);
    document.body.append(node);
    node.scrollIntoView();
}


