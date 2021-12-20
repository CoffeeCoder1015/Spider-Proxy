fetch("math.ans").then(function(r){
    return r.text();
}).then(function(text){
    document.getElementById("ans").innerHTML = text;
})