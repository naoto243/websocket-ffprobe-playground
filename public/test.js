

const newFfprobeApp = (f)=>{
  
  const self = {
    
    files : {
      target : f,
    },
    start  : ()=>{
      
      console.log(f);
        //fetch(`/open/${f.name}`)
      var loc = window.location;
      const url = `ws://${loc.host}/ws/start/${f.name}?size=${f.size}&file_type=${f.type}`;
      const ws = new WebSocket(url);
  
      ws.onopen = () =>  {
        console.log('Connected FFPROBE');
      };
  
      ws.onmessage = (evt) =>{
        
        const data = JSON.parse(evt.data);
        console.log(data);
  
        var reader = new FileReader();
  
        let blob = f.slice(data.start_byte, data.end_byte);
  
        reader.onload = function(e) {
          var buf = e.target.result;
          ws.send(buf);
        };
  
        reader.onerror = function(e) {
          ws.send(e);
        };
  
        //reader.readAsBinaryString(blob);
        
        reader.readAsArrayBuffer(blob)
      }
      
    },
  };
  
  return self;
  
};



