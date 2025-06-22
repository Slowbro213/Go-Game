//Deprecated!!!
//decode.js
const decoder = new TextDecoder()

export function decode(buf) {
  const view = new DataView(buf);
  let offset = 0;

  // Step 1: Extract message type
  const typeLen = view.getUint32(offset, true);
  offset += 4;

  const typeBytes = new Uint8Array(buf, offset, typeLen);
  const messageType = new TextDecoder().decode(typeBytes);
  offset += typeLen;

  const TYPE_MAP = ["character", "enemy", "item"];

  const objects = [];
  while (offset < buf.byteLength) {
    const id = view.getUint32(offset, true);
    offset += 4;

    const typeCode = view.getUint8(offset);
    offset += 1;

    const x = view.getFloat32(offset, true);
    offset += 4;

    const y = view.getFloat32(offset, true);
    offset += 4;

    objects.push({
      id,
      type: TYPE_MAP[typeCode] || "unknown",
      position: { x, y },
      children: []
    });
  }

  return {
    type: messageType,
    data: objects
  };
}

