const endpoint = "localhost:8080/ws"
const ws = new WebSocket("ws://" + endpoint)
const params = new URLSearchParams(window.location.search)

let username = params.get("user")
let room = params.get("room")
let roomsSetup = false

const appendMessage = (msg) => {
	const messageArea = document.querySelector("[data-message-area]")
	const newMessage = `
				<div class="flex flex-col w-full">
				<div class="text-blue-500">${msg.Data.sender} [${msg.Data.senderId}]</div>
					<div>${msg.Data.message}</div>
				</div>
			`
	messageArea.insertAdjacentHTML("beforeend", newMessage)
	messageArea.scrollTop = messageArea.scrollHeight
}

const onReceiveMessage = (event) => {
	const message = event.data
	const msgObj = JSON.parse(message)
	if (
		msgObj.MsgType === "pm" ||
		msgObj.MsgType === "broadcast" ||
		msgObj.MsgType === "local"
	) { appendMessage(msgObj) }
	else if (msgObj.MsgType === "status") setupRooms(msgObj)
	// health check is discarded
}

const onSendMessage = (ev) => {
	ev.preventDefault()
	ev.stopPropagation()
	const input = document.querySelector("#msgbox")

	const message = JSON.stringify({
		"MsgType": "message",
		"Data": {
			"sender": username,
			"room": room,
			"message": input.value
		}
	})
	ws.send(message)
	input.value = ""
	return false
}

const onOpen = (event) => {
	const header = {
		"msgType": "setup",
		"data": {
			"user": username,
			"room": room
		}
	}
	ws.send(JSON.stringify(header))
}

const setupName = () => {
	const nameDiv = document.querySelector("[data-username]")
	nameDiv.innerText = username
}

const setupUsers = () => {
	const userToggle = document.querySelector("#show-user-list")
	const userDiv = document.querySelector("[data-user-list]")
	userToggle.addEventListener("change", (ev) => {
		if(ev.target.checked) {
			userDiv.classList.remove("hidden")	
		}
		else {
			userDiv.classList.add("hidden")
		}
	})	
}

const getRooms = () => {
	const infoMsg = {
		"msgType": "listRooms",
		"data": {}
	}
	ws.send(JSON.stringify(infoMsg))
}

const setupRooms = (msgObj) => {
	const rooms = JSON.parse(msgObj.Data.rooms)
	const selectBox = document.querySelector("[data-room-selector]")
	selectBox.innerHTML = ""
	for (r of rooms) {
		let append = ''
		if (room === r) append = "selected"
		selectBox.insertAdjacentHTML('beforeend', `<option value='${r}' ${append}>${r}</option>`)
	}
	selectBox.insertAdjacentHTML('beforeend', `<option value='newChannel'>Create new channel...</option>`)
	if (!roomsSetup) {
		roomsSetup = true
		selectBox.addEventListener('change', (ev) => {

			if (ev.target.value === "newChannel") {
				let r = window.prompt("Enter the name of the new room.", "gen")
				if (r === null || r === "") return
				room = r
			}
			else room = ev.target.value

			const header = {
				"msgType": "setup",
				"data": {
					"user": username,
					"room": room
				}
			}
			ws.send(JSON.stringify(header))
		})
	}
}

window.addEventListener("load", () => {
	const form = document.querySelector("#sendmessage")
	form.addEventListener("submit", onSendMessage)
	setupName()
	setupUsers()
	getRooms()
	window.setInterval(getRooms, 10 * 1000)
})

ws.addEventListener("message", onReceiveMessage)
ws.addEventListener("open", onOpen)

