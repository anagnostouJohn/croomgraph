
"use client"
import { useRouter, useSearchParams } from 'next/navigation';
import axios from 'axios';
import Graph from '@/component/graphs/graphs';
import * as React from 'react';
import "./page.scss"


const RoomPage = ({ params }: { params: { room: string } }) => {
    const router = useRouter();
    // const { room } = params;   
    const [room, setRoom] = React.useState("")
    const [dummyData, setDummyData] = React.useState<any[]>([]);
    const searchParams = useSearchParams();
    const indexPlaceStr = searchParams.get('index');
    const indexPlace = indexPlaceStr ? parseInt(indexPlaceStr, 0) : 0;
    const list: string[] = [];


    interface DataType {
        [key: string]: any;  // Define your actual data structure here
    }

    const feachData = () => {
        axios.get("http://192.168.23.61:8080/data").then(res => {
            setDummyData(res.data["data"][indexPlace]["SensorsData"]);
            console.log(res.data["data"][indexPlace]);
            console.log("HEllo");
            setRoom(res.data["data"][indexPlace]["Room"])

        })
    }

    React.useEffect(() => {


        feachData();

        const interval = setInterval(feachData, 10000);
        return () => clearInterval(interval);
    }, [])


    const [selectedValue, setSelectedValue] = React.useState('1');

    const handleChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
        setSelectedValue(e.target.value);
        axios.post("http://192.168.23.61:8080/change", { value: e.target.value }).then(res => {
            console.log(res.status)

            feachData();
        })
    };
    return (<>
        <p className={"RoomP"}>{room}</p>
        <h1>Choose an Option</h1>
        <select value={selectedValue} onChange={handleChange}>
            <option value="1">Hour</option>
            <option value="2">Day</option>
            <option value="3">Week</option>
            <option value="4">Mounth</option>
        </select>

        <div className={"AllGraphs"}>
            {dummyData.map((key, value) => {
                return (
                    <>
                        {dummyData[value]["Temperature"].length == 0 && dummyData[value]["Humidity"].length == 0 ? <></> : <p className={"SensorP"}> Sensor : {key["Sensor"]}</p>}
                        <div className={"SensorGraphs"}>
                            {dummyData[value]["Temperature"].length == 0 ? <></> : <><div className={"TheGraph"}> <p className={"IndicatorP"}> Temperature </p>  <Graph data={dummyData[value]["Temperature"]} /></div> </>}
                            {dummyData[value]["Humidity"].length == 0 ? <></> : <><div className={"TheGraph"}> <p className={"IndicatorP"}> Humidity </p>  <Graph data={dummyData[value]["Humidity"]} /></div> </>}
                        </div>
                    </>

                )
            })
            }
        </div>


    </>)
}

export default RoomPage;