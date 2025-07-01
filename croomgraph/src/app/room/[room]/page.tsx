
"use client"
// import { useRouter, useSearchParams } from 'next/navigation';
import { useSearchParams } from 'next/navigation';
import { useState } from 'react';
import { useRouter } from "next/router";
import Calendar from 'react-calendar';
import axios from 'axios';
import Graph from '@/component/graphs/graphs';
import * as React from 'react';
import "./page.scss"

interface RoomPageProps {
    params: { room: string };
    searchParams?: { [key: string]: string | string[] | undefined };
}

// const RoomPage = ({ params }: { params: { room: string } }) => {
const RoomPage = ({ params, searchParams }: RoomPageProps) => {
    // const router = useRouter();
    // const router = useRouter();
    // const { room } = params;   
    const [room, setRoom] = React.useState("")
    const [readyForNext, setReadyForNext] = React.useState(true)
    const [sensorData, setSensorData] = React.useState<any[]>([]);
    const roomName = params.room;

    const indexPlace = Array.isArray(searchParams?.index)
        ? searchParams.index[0]
        : searchParams?.index;

    console.log(roomName, indexPlace, "AAAAAAAAAA<<<<<<<<<<<<<<<<<<<")


    let SendOnlyOnce = true
    let stop = false
    type ValuePiece = Date | null;

    type Value = ValuePiece | [ValuePiece, ValuePiece];
    const [selectedValue, setSelectedValue] = React.useState('1');
    const [valueCalendar, setValueCalendar] = useState<number>(0);
    const feachData = (num: number) => {

        if (readyForNext) {
            if (!stop) {
                setReadyForNext(false)
                //TODO CHHANGE TO ENV VALUE
                axios.get("http://192.168.23.61:8080" + "/data", { params: { Room: roomName } }).then(res => {
                    console.log(res, "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAL<MSAMAMDOASMDOMASDOMASMD")
                    setSensorData(res.data["data"]["SensorsData"])
                    // console.log
                    // setDummyData(res.data["data"][indexPlace]["SensorsData"]);
                    // console.log(res.data["data"][indexPlace]);
                    // console.log("HEllo");
                    // setRoom(res.data["data"][indexPlace]["Room"])
                    // setReadyForNext(true)
                });
            }
        }
    }

    React.useEffect(() => {
        const intervalId = setInterval(() => {
            feachData(0)
        }, 1000); // 1000 ms = 1 second
        return () => {
            clearInterval(intervalId);
        };

    }, [])


    const ChnageSelectedValueServer = (val: string) => {
        //TODO ADD global variable 
        axios.post("http://192.168.23.61:8080" + "/change", { value: val }).then(responce => {
            console.log(responce)
            feachData(0);
            SendOnlyOnce = true
        })


    }




    const handleCalendarChange = (newValue: Date, event: React.MouseEvent<HTMLButtonElement>) => {
        if (!newValue) {
            console.error('Selected value is+ null');
            return;
        }
        if (newValue instanceof Date) {
            setValueCalendar(newValue.getTime() / 1000);
            console.log(valueCalendar, "A!!A!!!QA!AA!A!A!")
            feachData(newValue.getTime() / 1000)
            // console.log('Selected date:', newValue.getTime()/1000);
        }
    };


    return (<>
        <p className={"RoomP"}>{room}</p>

        <div className={"AllGraphs"}>

            {roomName}

            {sensorData.map((item, index) => {
                console.log(item, "ASDASDSAAAOAOAOAOOAOAOAO");

                return (
                    <>
                    
                        {
                            (!item["Tc"]?.length && !item["h"]?.length) ? null :
                                <p className={"SensorP"}> Sensor : {item["Lab"]}</p>
                        }

                        <div className={"SensorGraphs"}>
                            {
                                item["Tc"]?.length ?
                                    <div className={"TheGraph"}>
                                        <p className={"IndicatorP"}> Temperature </p>
                                        <Graph data={item["Tc"]} selected={selectedValue} />
                                    </div>
                                    : null
                            }

                            {
                                item["H"]?.length ?
                                    <div className={"TheGraph"}>
                                        <p className={"IndicatorP"}> Humidity </p>
                                        <Graph data={item["H"]} selected={selectedValue} />
                                    </div>
                                    : null
                            }
                        </div>
                    </>
                );
            })}
        </div>
    </>)
}

export default RoomPage;









// React.useEffect(() => {
//     console.log(selectedValue, "AAAAAA!!!!!!!@!@!@!@!@$$$$$$$$$$$$")
//     if (selectedValue === "1") {
//         console.log("MESA STO 1 <<<<<<<<<<<<<<<<<<<")
//         //TODO ADD global variable
//         axios.post("http://192.168.23.61:8080"+"/change",  { value: "1" }).then(responce =>{
//             console.log(responce)
//             feachData(0);
//             SendOnlyOnce = true
//         })

//     } else if (selectedValue === "2") {
//         if (SendOnlyOnce) {
//             console.log("MESA", selectedValue, valueCalendar,  "ASSA<<<<<<<<<<<<<<<<<<<<<<<<<<")
//             //TODO ADD global variable
//             axios.post("http://192.168.23.61:8080"+"/change",  { value: "2" }).then(responce =>{
//                 console.log(responce, "ASAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA!!!!!!!!!!!!!!!!!!!!!!!!!!!")
//                 feachData(valueCalendar);
//                 SendOnlyOnce = false
//             })
//         }
//     }
//     const interval = setInterval(feachData, 10000);
//     return () => clearInterval(interval);
// // }, [valueCalendar, selectedValue])

// React.useEffect(() => {
//     if (selectedValue !== "1"){
//         return;
//     }
//     console.log(selectedValue, "AAAAAA!!!!!!!@!@!@!@!@$$$$$$$$$$$$")
//     if (selectedValue === "1") {
//         console.log("MESA STO 1 <<<<<<<<<<<<<<<<<<<")
//         //TODO ADD global variable
//         axios.post("http://192.168.23.61:8080" + "/change", { value: "1" }).then(responce => {
//             console.log(responce)
//             feachData(0);
//             SendOnlyOnce = true
//         })
//         const interval = setInterval(feachData, 10000);
//         return () => clearInterval(interval);
//     }

// }, [stop]);


// else if (selectedValue === "2") {
//     if (SendOnlyOnce) {
//         console.log("MESA", selectedValue, valueCalendar,  "ASSA<<<<<<<<<<<<<<<<<<<<<<<<<<")
//         //TODO ADD global variable
//         axios.post("http://192.168.23.61:8080"+"/change",  { value: "2" }).then(responce =>{
//             console.log(responce, "ASAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA!!!!!!!!!!!!!!!!!!!!!!!!!!!")
//             feachData(valueCalendar);
//             SendOnlyOnce = false
//         })
//     }
// }
// const handleSelectChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
//     setSelectedValue(e.target.value);
//     console.log("SADASDASDAKA!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
//     if (e.target.value == "1") {
//         setValueCalendar(0)
//         ChnageSelectedValueServer(e.target.value)
//         stop=false
//     } else if (e.target.value == "2") {
//         console.log("MEDAS MEESA MESA")
//         const now = new Date(); // Current date and time
//         const startOfDay = new Date(now.getFullYear(), now.getMonth(), now.getDate()); // Start of the day
//         const epochStartOfDay = startOfDay.getTime(); // Milliseconds since epoch
//         setValueCalendar(epochStartOfDay / 1000)
//         ChnageSelectedValueServer(e.target.value)
//         stop=true
//     }
//     ////TODO CHHANGE TO ENV VALUE
//     // axios.post("http://192.168.23.61:8080" + "/change", { value: e.target.value }).then(res => {
//     //     console.log(res.status)

//     //     feachData();
//     // })
// };