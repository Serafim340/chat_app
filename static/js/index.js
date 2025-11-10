async function loadRooms() {
  try {
    const res = await fetch("/rooms");
    const rooms = await res.json();

    const list = document.getElementById("roomList");
    list.innerHTML = "";

    rooms.forEach(r => {
      const btn = document.createElement("button");
      btn.textContent = `Join: ${r}`;
      btn.onclick = () => {
        const nick = prompt("Enter your nickname:");
        if (!nick) return;
        window.location.href = `/chat?room=${r}&nick=${nick}`;
      };
      list.appendChild(btn);
    });

  } catch (err) {
    console.error("Failed to load rooms:", err);
  }
}

loadRooms();
