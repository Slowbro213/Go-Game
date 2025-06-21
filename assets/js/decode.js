//Deprecated!!!
//decode.js
export function decode(buf) {
  const view = new DataView(buf);
  let offset = 0;

  // Step 1: Extract message type
  const typeLen = view.getUint32(offset, true);
  offset += 4;

  const typeBytes = new Uint8Array(buf, offset, typeLen);
  const messageType = new TextDecoder().decode(typeBytes);
  offset += typeLen;

  // Step 2: Decode all Concrete objects from the remaining buffer
  const objects = [];

  while (offset < buf.byteLength) {
    // ID: 4 bytes
    const id = view.getUint32(offset, true);
    offset += 4;

    // Game object type string: assume fixed (e.g. "character")
    const objTypeLen = "character".length;
    const objTypeBytes = new Uint8Array(buf, offset, objTypeLen);
    const objType = new TextDecoder().decode(objTypeBytes);
    offset += objTypeLen;

    // Position.X: 4 bytes
    const posX = view.getFloat32(offset, true);
    offset += 4;

    // Position.Y: 4 bytes
    const posY = view.getFloat32(offset, true);
    offset += 4;


    objects.push({
      id,
      type: objType,
      position: { x: posX, y: posY }
    });
  }

  return {
    type: messageType,
    data: objects
  };
}
