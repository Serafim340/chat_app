const params = new URLSearchParams(window.location.search);
const room = params.get("room");
const nick = params.get("nick");

if (!room || !nick) {
  alert("Missing room or nickname. Redirecting…");
  window.location.href = "/";
}

document.getElementById("roomTitle").textContent = room;

const socket = new WebSocket(`ws://${location.host}/room?room=${room}&nick=${nick}`);

socket.onmessage = (event) => {
  let data;
  try {
    data = JSON.parse(event.data);
  } catch {
    console.warn("Invalid JSON:", event.data);
    return;
  }

  if (data.type === "users") {
    updateUserPopup(data.list);
    return;
  }

  if (data.type === "message") {
    addMessage(data.name, data.message);
  }
};

function addMessage(name, message) {
  const c = document.createElement("div");
  c.classList.add("message-container");

  const u = document.createElement("div");
  u.classList.add("username");
  u.textContent = name;

  const m = document.createElement("div");
  m.classList.add("message");
  m.textContent = message;

  c.appendChild(u);
  c.appendChild(m);
  const messagesDiv = document.getElementById("messages");
  messagesDiv.appendChild(c);
  messagesDiv.scrollTop = messagesDiv.scrollHeight;
}

function updateUserPopup(users) {
  const popup = document.getElementById("usersPopup");
  popup.innerHTML = "<b>Online:</b><br>";

  users.forEach(u => {
    const div = document.createElement("div");
    div.textContent = u;
    popup.appendChild(div);
  });
}

document.getElementById("usersBtn").onclick = () => {
  document.getElementById("usersPopup").classList.toggle("hidden");
};

function sendMessage() {
  const input = document.getElementById("msg");
  if (input.value.trim()) {
    socket.send(input.value);
    input.value = "";
  }
}

document.getElementById("sendBtn").onclick = sendMessage;
document.getElementById("msg").addEventListener("keyup", e => {
  if (e.key === "Enter") sendMessage();
});
